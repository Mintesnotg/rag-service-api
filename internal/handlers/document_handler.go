package handlers

import (
	"errors"
	"net/http"
	"strings"

	"go-api/internal/services"

	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	documentService services.DocumentService
}

type createDocumentRequest struct {
	DocName        string `form:"doc_name" binding:"required"`
	DocDescription string `form:"doc_description"`
	Category       string `form:"category" binding:"required"`
}

type updateDocumentRequest struct {
	DocName        string `form:"doc_name" binding:"required"`
	DocDescription string `form:"doc_description"`
	Category       string `form:"category" binding:"required"`
}

func NewDocumentHandler(documentService services.DocumentService) *DocumentHandler {
	return &DocumentHandler{documentService: documentService}
}

// CreateDocument godoc
// @Summary Create document
// @Tags Documents
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param doc_name formData string true "Document title"
// @Param doc_description formData string false "Document description"
// @Param category formData string true "Category name (auto from page context)"
// @Param file formData file true "Document file"
// @Success 201 {object} services.DocumentDTO
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/documents [post]
func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	var req createDocumentRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document input"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	document, err := h.documentService.CreateDocument(c.Request.Context(), services.CreateDocumentInput{
		DocName:        req.DocName,
		DocDescription: req.DocDescription,
		CategoryName:   req.Category,
		File:           file,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrDocumentInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document input"})
		case errors.Is(err, services.ErrDocumentCategory):
			c.JSON(http.StatusNotFound, gin.H{"error": "document category not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create document"})
		}
		return
	}

	c.JSON(http.StatusCreated, document)
}

// ListDocuments godoc
// @Summary List documents by category
// @Tags Documents
// @Security BearerAuth
// @Produce json
// @Param category query string true "Category name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/documents [get]
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	category := strings.TrimSpace(c.Query("category"))
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category is required"})
		return
	}

	documents, err := h.documentService.ListDocuments(c.Request.Context(), category)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrDocumentInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "category is required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch documents"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": documents,
	})
}

// UpdateDocument godoc
// @Summary Update document
// @Tags Documents
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "document id"
// @Param doc_name formData string true "Document title"
// @Param doc_description formData string false "Document description"
// @Param category formData string true "Category name (auto from page context)"
// @Param file formData file false "Replacement file"
// @Success 200 {object} services.DocumentDTO
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/documents/{id} [put]
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document id is required"})
		return
	}

	var req updateDocumentRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document input"})
		return
	}

	file, _ := c.FormFile("file")

	document, err := h.documentService.UpdateDocument(c.Request.Context(), id, services.UpdateDocumentInput{
		DocName:        req.DocName,
		DocDescription: req.DocDescription,
		CategoryName:   req.Category,
		File:           file,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrDocumentInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document input"})
		case errors.Is(err, services.ErrDocumentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		case errors.Is(err, services.ErrDocumentCategory):
			c.JSON(http.StatusNotFound, gin.H{"error": "document category not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update document"})
		}
		return
	}

	c.JSON(http.StatusOK, document)
}

// DeleteDocument godoc
// @Summary Soft delete document (set status inactive)
// @Tags Documents
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 204 "deleted"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/documents/{id} [delete]
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document id is required"})
		return
	}

	if err := h.documentService.DeleteDocument(c.Request.Context(), id); err != nil {
		switch {
		case errors.Is(err, services.ErrDocumentInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document input"})
		case errors.Is(err, services.ErrDocumentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete document"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GetDocumentDownloadURL godoc
// @Summary Get temporary download URL for a document
// @Tags Documents
// @Security BearerAuth
// @Produce json
// @Param id path string true "document id"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/documents/{id}/download [get]
func (h *DocumentHandler) GetDocumentDownloadURL(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document id is required"})
		return
	}

	url, err := h.documentService.GetDownloadURL(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrDocumentInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document input"})
		case errors.Is(err, services.ErrDocumentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate download url"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}
