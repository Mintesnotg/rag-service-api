package handlers

import (
	"net/http"
	"strconv"

	"go-api/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PermissionHandler struct {
	permissionService services.PermissionService
}

type PermissionRequest struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}

type CreatePermissionRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdatePermissionRequest struct {
	Name string `json:"name" binding:"required"`
}

func NewPermissionHandler(permissionService services.PermissionService) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService}
}

// ListPermissions godoc
// @Summary List permissions
// @Tags Permissions
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search by name"
// @Success 200 {object} map[string]interface{}
// @Router /api/permissions [get]
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	perms, total, err := h.permissionService.ListPermissions(search, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      perms,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"search":    search,
	})
}

// CreatePermission godoc
// @Summary Create a permission
// @Tags Permissions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body CreatePermissionRequest true "create permission"
// @Success 201 {object} services.PermissionDTO
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/permissions [post]
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	perm, err := h.permissionService.CreatePermission(req.Name)
	if err != nil {
		if services.IsConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "permission name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create permission"})
		return
	}

	c.JSON(http.StatusCreated, perm)
}

// UpdatePermission godoc
// @Summary Update a permission
// @Tags Permissions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "permission id"
// @Param data body UpdatePermissionRequest true "update permission"
// @Success 200 {object} services.PermissionDTO
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/permissions/{id} [put]
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")
	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	perm, err := h.permissionService.UpdatePermission(id, req.Name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "permission not found"})
			return
		}
		if services.IsConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "permission name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update permission"})
		return
	}

	c.JSON(http.StatusOK, perm)
}

// DeletePermission godoc
// @Summary Soft delete a permission
// @Tags Permissions
// @Security BearerAuth
// @Param id path string true "permission id"
// @Success 204 "deleted"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/permissions/{id} [delete]
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	err := h.permissionService.DeletePermission(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "permission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete permission"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ResolvePermissions godoc
// @Summary Get permissions for the given role IDs
// @Tags Permissions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body PermissionRequest true "role ids"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/permissions/resolve [post]
func (h *PermissionHandler) ResolvePermissions(c *gin.Context) {
	var req PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.RoleIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role_ids are required"})
		return
	}

	perms, err := h.permissionService.GetPermissionsByRoleIDs(req.RoleIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": perms})
}
