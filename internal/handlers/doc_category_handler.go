package handlers

import (
	"net/http"
	"time"

	"go-api/internal/enums"
	"go-api/internal/services"

	"github.com/gin-gonic/gin"
)

type DocCategoryHandler struct {
	service services.DocCategoryService
}

type CreateDocCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateDocCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type DocCategoryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewDocCategoryHandler(service services.DocCategoryService) *DocCategoryHandler {
	return &DocCategoryHandler{service: service}
}

func toDocCategoryResponse(id, name, description string, status any, createdAt, updatedAt time.Time) DocCategoryResponse {
	return DocCategoryResponse{
		ID:          id,
		Name:        name,
		Description: description,
		Status:      stringifyStatus(status),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func stringifyStatus(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case enums.RecordStatus:
		return string(t)
	default:
		return ""
	}
}

// CreateDocCategory godoc
// @Summary Create a document category
// @Tags Document Categories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body CreateDocCategoryRequest true "document category"
// @Success 201 {object} DocCategoryResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/doc-categories [post]
func (h *DocCategoryHandler) CreateDocCategory(c *gin.Context) {
	var req CreateDocCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	category, err := h.service.CreateDocCategory(req.Name, req.Description)
	if err != nil {
		switch err {
		case services.ErrDocCategoryInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create document category"})
		}
		return
	}

	c.JSON(http.StatusCreated, toDocCategoryResponse(
		category.ID,
		category.Name,
		category.Description,
		category.Status,
		category.CreatedAt,
		category.UpdatedAt,
	))
}

// GetAllDocCategories godoc
// @Summary Get all document categories
// @Tags Document Categories
// @Security BearerAuth
// @Produce json
// @Success 200 {array} DocCategoryResponse
// @Failure 500 {object} map[string]string
// @Router /api/doc-categories [get]
func (h *DocCategoryHandler) GetAllDocCategories(c *gin.Context) {
	categories, err := h.service.GetAllDocCategory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch document categories"})
		return
	}
	resp := make([]DocCategoryResponse, 0, len(categories))
	for _, cat := range categories {
		resp = append(resp, toDocCategoryResponse(
			cat.ID,
			cat.Name,
			cat.Description,
			cat.Status,
			cat.CreatedAt,
			cat.UpdatedAt,
		))
	}
	c.JSON(http.StatusOK, resp)
}

// GetDocCategoryByID godoc
// @Summary Get a document category by ID
// @Tags Document Categories
// @Security BearerAuth
// @Produce json
// @Param id path string true "category id"
// @Success 200 {object} DocCategoryResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/doc-categories/{id} [get]
func (h *DocCategoryHandler) GetDocCategoryByID(c *gin.Context) {
	id := c.Param("id")

	category, err := h.service.GetDocCategoryByID(id)
	if err != nil {
		switch err {
		case services.ErrDocCategoryInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		case services.ErrDocCategoryNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "document category not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch document category"})
		}
		return
	}

	c.JSON(http.StatusOK, toDocCategoryResponse(
		category.ID,
		category.Name,
		category.Description,
		category.Status,
		category.CreatedAt,
		category.UpdatedAt,
	))
}

// UpdateDocCategory godoc
// @Summary Update a document category
// @Tags Document Categories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "category id"
// @Param data body UpdateDocCategoryRequest true "document category"
// @Success 200 {object} DocCategoryResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/doc-categories/{id} [put]
func (h *DocCategoryHandler) UpdateDocCategory(c *gin.Context) {
	id := c.Param("id")

	var req UpdateDocCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	category, err := h.service.UpdateDocCategory(id, req.Name, req.Description)
	if err != nil {
		switch err {
		case services.ErrDocCategoryInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		case services.ErrDocCategoryNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "document category not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update document category"})
		}
		return
	}

	c.JSON(http.StatusOK, toDocCategoryResponse(
		category.ID,
		category.Name,
		category.Description,
		category.Status,
		category.CreatedAt,
		category.UpdatedAt,
	))
}

// DeleteDocCategory godoc
// @Summary Soft delete a document category (set status inactive)
// @Tags Document Categories
// @Security BearerAuth
// @Produce json
// @Param id path string true "category id"
// @Success 204 {string} string "no content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/doc-categories/{id} [delete]
func (h *DocCategoryHandler) DeleteDocCategory(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteDocCategory(id); err != nil {
		switch err {
		case services.ErrDocCategoryInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		case services.ErrDocCategoryNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "document category not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete document category"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

