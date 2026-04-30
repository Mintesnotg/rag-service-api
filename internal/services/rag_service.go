package services

import (
	"context"
	"encoding/json"
	"errors"
	"go-api/internal/enums"
	docmodels "go-api/internal/models/doc-category"
	ragmodels "go-api/internal/models/rag"
	"go-api/internal/repositories"
	"go-api/internal/storage"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"
)

var ErrRAGInvalidInput = errors.New("invalid rag input")

type RAGService interface {
	IndexDocument(ctx context.Context, document *docmodels.Document) error
	RemoveDocument(ctx context.Context, documentID string) error
	Query(ctx context.Context, input QueryInput) (*QueryResponse, error)
}

type QueryInput struct {
	Question string
	Category string
	TopK     int
}

type QueryResponse struct {
	Answer   string   `json:"answer"`
	Sources  []string `json:"sources"`
	Contexts []string `json:"contexts"`
}

const responseFormatGuide = `Enhanced Prompt (Structured and user-friendly):
Generate responses that are well-formatted, structured, and professional. Follow these strict formatting and usability rules.

1. Text formatting rules
- Convert markdown-style syntax into clean formatted text.
- "**text**" should be rendered as bold text.
- Lines starting with "*" should be converted into proper bullet points.
- Do not display raw symbols like "**" or "*" in the final output.

2. Content structure
- Organize responses with clear section headings and subheadings.
- Group related information logically.
- Maintain proper spacing between sections.
- Keep paragraphs separated and indentation consistent.

3. Lists and flow
- Use bullet points for unordered information.
- Use numbered lists for step-by-step instructions.
- Keep a smooth logical flow: Introduction -> Details -> Summary (when applicable).

4. Writing style
- Use professional language.
- Keep it clear, concise, and easy to understand.
- Avoid redundancy and unnecessary clutter.`

type ragService struct {
	documentRepo repositories.DocumentRepository
	ragRepo      repositories.RAGChunkRepository
	fileStore    storage.ObjectStorage
	chunker      Chunker
	extractor    MultiExtractor
	embedder     Embedder
	llm          LLM
}

func NewRAGService(
	documentRepo repositories.DocumentRepository,
	ragRepo repositories.RAGChunkRepository,
	fileStore storage.ObjectStorage,
	chunker Chunker,
	extractor MultiExtractor,
	embedder Embedder,
	llm LLM,
) RAGService {
	if extractor == nil {
		extractor = NewExtractorChain(NewPlainTextExtractor())
	}
	return &ragService{
		documentRepo: documentRepo,
		ragRepo:      ragRepo,
		fileStore:    fileStore,
		chunker:      chunker,
		extractor:    extractor,
		embedder:     embedder,
		llm:          llm,
	}
}

func (s *ragService) IndexDocument(ctx context.Context, document *docmodels.Document) error {
	if document == nil || strings.TrimSpace(document.ID) == "" || strings.TrimSpace(document.FileURL) == "" {
		log.Printf("rag: invalid index document input: %+v", document)
		return ErrRAGInvalidInput
	}

	s.updateProcessingStatus(document.ID, enums.ProcessingRunning)
	reader, err := s.fileStore.Download(ctx, document.FileURL)
	if err != nil {
		s.updateProcessingStatus(document.ID, enums.ProcessingFailed)
		log.Printf("rag: failed to download document file document_id=%s file_url=%s err=%v", document.ID, document.FileURL, err)
		return err

	}
	defer reader.Close()

	contentBytes, err := io.ReadAll(io.LimitReader(reader, 10*1024*1024))
	if err != nil {
		s.updateProcessingStatus(document.ID, enums.ProcessingFailed)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			log.Printf("rag: indexing canceled while reading document bytes document_id=%s err=%v", document.ID, err)
		} else {
			log.Printf("rag: failed to read document bytes document_id=%s err=%v", document.ID, err)
		}
		return err
	}

	content, err := s.extractor.Extract(
		ctx,
		filepath.Base(strings.TrimSpace(document.FileURL)),
		strings.TrimSpace(document.FileType),
		contentBytes,
	)
	if err != nil {
		s.updateProcessingStatus(document.ID, enums.ProcessingFailed)
		log.Printf("rag: failed to extract text document_id=%s err=%v", document.ID, err)
		return err
	}
	content = strings.TrimSpace(content)
	if content == "" {
		s.updateProcessingStatus(document.ID, enums.ProcessingFailed)
		log.Printf("rag: extracted empty text document_id=%s", document.ID)
		return errors.New("document has no text content")
	}

	chunksText := s.chunker.Chunk(content)
	chunks := make([]ragmodels.Chunk, 0, len(chunksText))
	embeddings := make([][]float64, 0, len(chunksText))
	chunkCreatedAt := time.Now().UTC()
	for i, chunk := range chunksText {
		vector, embedErr := s.embedder.Embed(ctx, chunk)
		if embedErr != nil {
			s.updateProcessingStatus(document.ID, enums.ProcessingFailed)
			log.Printf("rag: embedding failed document_id=%s chunk_index=%d err=%v", document.ID, i, embedErr)
			return embedErr
		}
		metadata, metadataErr := json.Marshal(map[string]interface{}{
			"doc_name":      document.DocName,
			"category_id":   document.CategoryID,
			"category_name": document.Category.Name,
		})
		if metadataErr != nil {
			log.Printf("rag: failed to marshal chunk metadata document_id=%s chunk_index=%d err=%v", document.ID, i, metadataErr)
		}
		chunks = append(chunks, ragmodels.Chunk{
			DocumentID: document.ID,
			ChunkIndex: i,
			Content:    chunk,
			Metadata:   string(metadata),
			CreatedAt:  chunkCreatedAt,
		})
		embeddings = append(embeddings, vector)
	}

	if err := s.ragRepo.ReplaceDocumentChunks(document.ID, chunks, embeddings); err != nil {
		s.updateProcessingStatus(document.ID, enums.ProcessingFailed)
		log.Printf("rag: failed to replace document chunks document_id=%s err=%v", document.ID, err)
		return err
	}
	s.updateProcessingStatus(document.ID, enums.ProcessingCompleted)
	return nil
}

func (s *ragService) RemoveDocument(ctx context.Context, documentID string) error {
	_ = ctx
	return s.ragRepo.DeleteByDocumentID(documentID)
}

func (s *ragService) Query(ctx context.Context, input QueryInput) (*QueryResponse, error) {
	question := strings.TrimSpace(input.Question)
	if question == "" {
		log.Printf("rag: query rejected because question is empty")
		return nil, ErrRAGInvalidInput
	}
	if input.TopK <= 0 {
		input.TopK = 5
	}

	queryEmbedding, err := s.embedder.Embed(ctx, question)
	if err != nil {
		log.Printf("rag: failed to embed query question=%q err=%v", question, err)
		return nil, err
	}
	results, err := s.ragRepo.SearchSimilar(queryEmbedding, input.TopK, strings.TrimSpace(input.Category))
	if err != nil {
		log.Printf("rag: failed to search similar chunks question=%q category=%q top_k=%d err=%v", question, input.Category, input.TopK, err)
		return nil, err
	}

	contexts := make([]string, 0, len(results))
	sourcesMap := map[string]struct{}{}
	for _, result := range results {
		contexts = append(contexts, result.Content)
		sourcesMap[result.DocumentID] = struct{}{}
	}

	answerPrompt := question + "\n\n" + responseFormatGuide
	answer, err := s.llm.GenerateAnswer(ctx, answerPrompt, contexts)
	if err != nil {
		log.Printf("rag: failed to generate answer question=%q contexts=%d err=%v", question, len(contexts), err)
		return nil, err
	}

	sources := make([]string, 0, len(sourcesMap))
	for id := range sourcesMap {
		sources = append(sources, id)
	}
	return &QueryResponse{
		Answer:   answer,
		Sources:  sources,
		Contexts: contexts,
	}, nil
}

func (s *ragService) updateProcessingStatus(documentID string, status enums.ProcessingStatus) {
	if err := s.documentRepo.UpdateProcessingStatus(documentID, status); err != nil {
		log.Printf("rag: failed to update processing status document_id=%s status=%s err=%v", documentID, status, err)
	}
}
