package handlers

import (
	"errors"
	"net/http"
	"strings"

	"go-api/internal/services"

	"github.com/gin-gonic/gin"
)

type RAGHandler struct {
	ragService      services.RAGService
	documentService services.DocumentService
}

type RAGQueryRequest struct {
	Question string `json:"question" binding:"required"`
	Category string `json:"category"`
	TopK     int    `json:"top_k"`
}

func NewRAGHandler(
	ragService services.RAGService,
	documentService services.DocumentService,
) *RAGHandler {
	return &RAGHandler{
		ragService:      ragService,
		documentService: documentService,
	}
}

// Query godoc
// @Summary Query indexed documents with RAG
// @Description Returns an answer and supporting contexts from indexed documents.
// @Tags RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Permission header string true "Permission claim" default(query_rag)
// @Param data body RAGQueryRequest true "RAG query payload"
// @Success 200 {object} services.QueryResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
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

// GetSourceURL godoc
// @Summary Resolve source document download URL
// @Description Returns a temporary presigned URL for a cited document source.
// @Tags RAG
// @Security BearerAuth
// @Produce json
// @Param X-Permission header string true "Permission claim" default(download_reference_docs)
// @Param id path string true "document id"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/rag/source/{id} [get]
func (h *RAGHandler) GetSourceURL(c *gin.Context) {
	if h.documentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "document service is not configured on this environment",
		})
		return
	}

	documentID := strings.TrimSpace(c.Param("id"))
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document id is required"})
		return
	}

	url, err := h.documentService.GetDownloadURL(c.Request.Context(), documentID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrDocumentInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		case errors.Is(err, services.ErrDocumentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not resolve source url"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}
