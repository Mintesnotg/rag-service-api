package repositories

import (
	"fmt"
	"strings"

	ragmodels "go-api/internal/models/rag"

	"gorm.io/gorm"
)

type RAGSearchResult struct {
	DocumentID string
	Content    string
	Score      float64
}

type RAGChunkRepository interface {
	ReplaceDocumentChunks(documentID string, chunks []ragmodels.Chunk, embeddings [][]float64) error
	DeleteByDocumentID(documentID string) error
	SearchSimilar(embedding []float64, topK int, categoryName string) ([]RAGSearchResult, error)
}

type ragChunkRepository struct {
	db *gorm.DB
}

func NewRAGChunkRepository(db *gorm.DB) RAGChunkRepository {
	return &ragChunkRepository{db: db}
}

func (r *ragChunkRepository) ReplaceDocumentChunks(documentID string, chunks []ragmodels.Chunk, embeddings [][]float64) error {
	if err := r.DeleteByDocumentID(documentID); err != nil {
		return err
	}
	if len(chunks) == 0 {
		return nil
	}

	for i := range chunks {
		if i >= len(embeddings) {
			break
		}
		if err := r.db.Exec(
			`INSERT INTO rag_chunks (document_id, chunk_index, content, metadata, embedding, created_at) VALUES (?, ?, ?, ?::jsonb, ?::vector, ?)`,
			documentID,
			chunks[i].ChunkIndex,
			chunks[i].Content,
			chunks[i].Metadata,
			vectorLiteral(embeddings[i]),
			chunks[i].CreatedAt,
		).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *ragChunkRepository) DeleteByDocumentID(documentID string) error {
	return r.db.Where("document_id = ?", documentID).Delete(&ragmodels.Chunk{}).Error
}

func (r *ragChunkRepository) SearchSimilar(embedding []float64, topK int, categoryName string) ([]RAGSearchResult, error) {
	if topK <= 0 {
		topK = 5
	}

	query := `
SELECT rc.document_id, rc.content, 1 - (rc.embedding <=> ?::vector) AS score
FROM rag_chunks rc
JOIN documents d ON d.id = rc.document_id
JOIN doc_categories dc ON dc.id = d.category_id
WHERE d.status = 'active' AND dc.status = 'active'
`
	args := []interface{}{vectorLiteral(embedding)}
	if strings.TrimSpace(categoryName) != "" {
		query += " AND LOWER(dc.name) = LOWER(?)"
		args = append(args, categoryName)
	}
	query += " ORDER BY rc.embedding <=> ?::vector LIMIT ?"
	args = append(args, vectorLiteral(embedding), topK)

	var out []RAGSearchResult
	if err := r.db.Raw(query, args...).Scan(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func vectorLiteral(embedding []float64) string {
	parts := make([]string, 0, len(embedding))
	for _, value := range embedding {
		parts = append(parts, fmt.Sprintf("%f", value))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
