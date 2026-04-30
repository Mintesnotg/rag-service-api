package handlers

import (
	"errors"
	"net/http"

	"go-api/internal/services"

	"github.com/gin-gonic/gin"
)

type RAGHandler struct {
	ragService services.RAGService
}

type RAGQueryRequest struct {
	Question string `json:"question" binding:"required"`
	Category string `json:"category"`
	TopK     int    `json:"top_k"`
}

func NewRAGHandler(ragService services.RAGService) *RAGHandler {
	return &RAGHandler{ragService: ragService}
}

// Query godoc
// @Summary Query indexed documents with RAG
// @Description Returns an answer and supporting contexts from indexed documents.
// @Tags RAG
// @Accept json
// @Produce json
// @Param data body RAGQueryRequest true "RAG query payload"
// @Success 200 {object} services.QueryResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/rag/query [post]
func (h *RAGHandler) Query(c *gin.Context) {
	if h.ragService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "rag service is not configured on this environment",
		})
		return
	}

	var req RAGQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rag query"})
		return
	}

	result, err := h.ragService.Query(c.Request.Context(), services.QueryInput{
		Question: req.Question,
		Category: req.Category,
		TopK:     req.TopK,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRAGInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "question is required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}
