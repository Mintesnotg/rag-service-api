package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

type Chunker interface {
	Chunk(text string) []string
}

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float64, error)
}

type LLM interface {
	GenerateAnswer(ctx context.Context, question string, contexts []string) (string, error)
}

type Extractor interface {
	CanHandle(fileName, contentType string) bool
	Extract(ctx context.Context, fileName, contentType string, payload []byte) (string, error)
}

type MultiExtractor interface {
	Extract(ctx context.Context, fileName, contentType string, payload []byte) (string, error)
}

type simpleChunker struct {
	chunkSize    int
	chunkOverlap int
}

func NewSimpleChunker(chunkSize, chunkOverlap int) Chunker {
	if chunkSize <= 0 {
		chunkSize = 1200
	}
	if chunkOverlap < 0 {
		chunkOverlap = 150
	}
	if chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 4
	}
	return &simpleChunker{chunkSize: chunkSize, chunkOverlap: chunkOverlap}
}

func (c *simpleChunker) Chunk(text string) []string {
	normalized := strings.TrimSpace(text)
	if normalized == "" {
		return nil
	}
	runes := []rune(normalized)
	var chunks []string
	step := c.chunkSize - c.chunkOverlap
	for start := 0; start < len(runes); start += step {
		end := start + c.chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		piece := strings.TrimSpace(string(runes[start:end]))
		if piece != "" {
			chunks = append(chunks, piece)
		}
		if end == len(runes) {
			break
		}
	}
	return chunks
}

type extractorChain struct {
	extractors []Extractor
}

func NewExtractorChain(extractors ...Extractor) MultiExtractor {
	return &extractorChain{extractors: extractors}
}

func (c *extractorChain) Extract(ctx context.Context, fileName, contentType string, payload []byte) (string, error) {
	_ = ctx
	for _, extractor := range c.extractors {
		if !extractor.CanHandle(fileName, contentType) {
			continue
		}
		text, err := extractor.Extract(ctx, fileName, contentType, payload)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(text) != "" {
			return normalizeExtractedText(text), nil
		}
	}
	return "", errors.New("no extractor found for file type")
}

type plainTextExtractor struct{}

func NewPlainTextExtractor() Extractor {
	return &plainTextExtractor{}
}

func (e *plainTextExtractor) CanHandle(fileName, contentType string) bool {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))
	ct := strings.ToLower(strings.TrimSpace(contentType))
	return ext == ".txt" || strings.Contains(ct, "text/plain")
}

func (e *plainTextExtractor) Extract(ctx context.Context, fileName, contentType string, payload []byte) (string, error) {
	_ = ctx
	_ = fileName
	_ = contentType
	return string(payload), nil
}

type pdfExtractor struct{}

func NewPDFExtractor() Extractor {
	return &pdfExtractor{}
}

func (e *pdfExtractor) CanHandle(fileName, contentType string) bool {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))
	ct := strings.ToLower(strings.TrimSpace(contentType))
	return ext == ".pdf" || strings.Contains(ct, "application/pdf")
}

func (e *pdfExtractor) Extract(ctx context.Context, fileName, contentType string, payload []byte) (string, error) {
	_ = ctx
	_ = fileName
	_ = contentType

	reader := bytes.NewReader(payload)
	pdfReader, err := pdf.NewReader(reader, int64(len(payload)))
	if err != nil {
		return "", err
	}
	var all strings.Builder
	totalPages := pdfReader.NumPage()
	for pageIndex := 1; pageIndex <= totalPages; pageIndex++ {
		page := pdfReader.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}
		text, extractErr := page.GetPlainText(nil)
		if extractErr != nil {
			continue
		}
		all.WriteString(text)
		all.WriteString("\n")
	}
	return all.String(), nil
}

type docxExtractor struct{}

func NewDOCXExtractor() Extractor {
	return &docxExtractor{}
}

func (e *docxExtractor) CanHandle(fileName, contentType string) bool {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))
	ct := strings.ToLower(strings.TrimSpace(contentType))
	return ext == ".docx" || strings.Contains(ct, "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
}

func (e *docxExtractor) Extract(ctx context.Context, fileName, contentType string, payload []byte) (string, error) {
	_ = ctx
	_ = fileName
	_ = contentType

	reader := bytes.NewReader(payload)
	zipReader, err := zip.NewReader(reader, int64(len(payload)))
	if err != nil {
		return "", err
	}
	for _, file := range zipReader.File {
		if file.Name != "word/document.xml" {
			continue
		}
		xmlFile, openErr := file.Open()
		if openErr != nil {
			return "", openErr
		}
		data, readErr := io.ReadAll(xmlFile)
		_ = xmlFile.Close()
		if readErr != nil {
			return "", readErr
		}

		xmlText := string(data)
		xmlText = strings.ReplaceAll(xmlText, "</w:p>", "\n")
		xmlText = strings.ReplaceAll(xmlText, "</w:tr>", "\n")
		re := regexp.MustCompile(`<[^>]+>`)
		return re.ReplaceAllString(xmlText, " "), nil
	}
	return "", errors.New("word/document.xml not found in docx")
}

func normalizeExtractedText(text string) string {
	lines := strings.Split(text, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, "\n")
}

type geminiClient struct {
	apiKey      string
	embedModel  string
	embedDim    int
	chatModel   string
	httpClient  *http.Client
	embedAPIURL string
	chatAPIURL  string
}

func NewGeminiClient() (Embedder, LLM, error) {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		return nil, nil, errors.New("GEMINI_API_KEY is required")
	}

	embedModel := strings.TrimSpace(os.Getenv("GEMINI_EMBED_MODEL"))
	if embedModel == "" {
		embedModel = "gemini-embedding-001"
	}
	embedModel = canonicalModelName(embedModel)

	embedDim := 768
	if rawDim := strings.TrimSpace(os.Getenv("GEMINI_EMBED_DIM")); rawDim != "" {
		parsed, parseErr := strconv.Atoi(rawDim)
		if parseErr == nil && parsed > 0 {
			embedDim = parsed
		}
	}

	chatModel := strings.TrimSpace(os.Getenv("GEMINI_CHAT_MODEL"))
	if chatModel == "" {
		chatModel = "gemini-1.5-flash"
	}
	chatModel = canonicalModelName(chatModel)

	client := &geminiClient{
		apiKey:      apiKey,
		embedModel:  embedModel,
		embedDim:    embedDim,
		chatModel:   chatModel,
		httpClient:  &http.Client{Timeout: 40 * time.Second},
		embedAPIURL: "https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s",
		chatAPIURL:  "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
	}
	return client, client, nil
}

func (g *geminiClient) Embed(ctx context.Context, text string) ([]float64, error) {
	reqBody := map[string]interface{}{
		"model": fmt.Sprintf("models/%s", g.embedModel),
		"content": map[string]interface{}{
			"parts": []map[string]string{{"text": text}},
		},
	}
	if g.embedDim > 0 {
		reqBody["outputDimensionality"] = g.embedDim
	}
	rawBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(g.embedAPIURL, g.embedModel, g.apiKey), bytes.NewReader(rawBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gemini embed failed: %s", string(body))
	}

	var parsed struct {
		Embedding struct {
			Values []float64 `json:"values"`
		} `json:"embedding"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Embedding.Values) == 0 {
		return nil, errors.New("empty embedding from gemini")
	}
	return parsed.Embedding.Values, nil
}

func (g *geminiClient) GenerateAnswer(ctx context.Context, question string, contexts []string) (string, error) {
	prompt := "Use only the provided context to answer. If not enough info, say so.\n\nContext:\n" +
		strings.Join(contexts, "\n---\n") +
		"\n\nQuestion: " + question

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{{"text": prompt}},
			},
		},
	}
	rawBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(g.chatAPIURL, g.chatModel, g.apiKey), bytes.NewReader(rawBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini chat failed: %s", string(body))
	}

	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty answer from gemini")
	}
	return parsed.Candidates[0].Content.Parts[0].Text, nil
}

func canonicalModelName(model string) string {
	model = strings.TrimSpace(model)
	model = strings.TrimPrefix(model, "models/")
	return strings.TrimSpace(model)
}
