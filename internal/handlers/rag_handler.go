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

type ragQueryRequest struct {
	Question string `json:"question" binding:"required"`
	Category string `json:"category"`
	TopK     int    `json:"top_k"`
}

func NewRAGHandler(ragService services.RAGService) *RAGHandler {
	return &RAGHandler{ragService: ragService}
}

func (h *RAGHandler) Query(c *gin.Context) {
	var req ragQueryRequest
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
