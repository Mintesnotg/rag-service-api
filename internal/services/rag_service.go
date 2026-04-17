package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"go-api/internal/enums"
	docmodels "go-api/internal/models/doc-category"
	ragmodels "go-api/internal/models/rag"
	"go-api/internal/repositories"
	"go-api/internal/storage"
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
		return ErrRAGInvalidInput
	}

	_ = s.documentRepo.UpdateProcessingStatus(document.ID, enums.ProcessingRunning)
	reader, err := s.fileStore.Download(ctx, document.FileURL)
	if err != nil {
		_ = s.documentRepo.UpdateProcessingStatus(document.ID, enums.ProcessingFailed)
		return err
	}
	defer reader.Close()

	contentBytes, err := io.ReadAll(io.LimitReader(reader, 10*1024*1024))
	if err != nil {
		_ = s.documentRepo.UpdateProcessingStatus(document.ID, enums.ProcessingFailed)
		return err
	}

	content, err := s.extractor.Extract(
		ctx,
		filepath.Base(strings.TrimSpace(document.FileURL)),
		strings.TrimSpace(document.FileType),
		contentBytes,
	)
	if err != nil {
		_ = s.documentRepo.UpdateProcessingStatus(document.ID, enums.ProcessingFailed)
		return err
	}
	content = strings.TrimSpace(content)
	if content == "" {
		_ = s.documentRepo.UpdateProcessingStatus(document.ID, enums.ProcessingFailed)
		return errors.New("document has no text content")
	}

	chunksText := s.chunker.Chunk(content)
	chunks := make([]ragmodels.Chunk, 0, len(chunksText))
	embeddings := make([][]float64, 0, len(chunksText))
	for i, chunk := range chunksText {
		vector, embedErr := s.embedder.Embed(ctx, chunk)
		if embedErr != nil {
			_ = s.documentRepo.UpdateProcessingStatus(document.ID, enums.ProcessingFailed)
			return embedErr
		}
		metadata, _ := json.Marshal(map[string]interface{}{
			"doc_name":      document.DocName,
			"category_id":   document.CategoryID,
			"category_name": document.Category.Name,
		})
		chunks = append(chunks, ragmodels.Chunk{
			DocumentID: document.ID,
			ChunkIndex: i,
			Content:    chunk,
			Metadata:   string(metadata),
		})
		embeddings = append(embeddings, vector)
	}

	if err := s.ragRepo.ReplaceDocumentChunks(document.ID, chunks, embeddings); err != nil {
		_ = s.documentRepo.UpdateProcessingStatus(document.ID, enums.ProcessingFailed)
		return err
	}
	_ = s.documentRepo.UpdateProcessingStatus(document.ID, enums.ProcessingCompleted)
	return nil
}

func (s *ragService) RemoveDocument(ctx context.Context, documentID string) error {
	_ = ctx
	return s.ragRepo.DeleteByDocumentID(documentID)
}

func (s *ragService) Query(ctx context.Context, input QueryInput) (*QueryResponse, error) {
	question := strings.TrimSpace(input.Question)
	if question == "" {
		return nil, ErrRAGInvalidInput
	}
	if input.TopK <= 0 {
		input.TopK = 5
	}

	queryEmbedding, err := s.embedder.Embed(ctx, question)
	if err != nil {
		return nil, err
	}
	results, err := s.ragRepo.SearchSimilar(queryEmbedding, input.TopK, strings.TrimSpace(input.Category))
	if err != nil {
		return nil, err
	}

	contexts := make([]string, 0, len(results))
	sourcesMap := map[string]struct{}{}
	for _, result := range results {
		contexts = append(contexts, result.Content)
		sourcesMap[result.DocumentID] = struct{}{}
	}

	answer, err := s.llm.GenerateAnswer(ctx, question, contexts)
	if err != nil {
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
