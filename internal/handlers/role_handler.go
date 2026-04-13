package handlers

import (
	"net/http"
	"strconv"

	"go-api/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleHandler struct {
	roleService services.RoleService
}

type CreateRoleRequest struct {
	Name          string   `json:"name" binding:"required"`
	PermissionIDs []string `json:"permission_ids"`
}

type UpdateRoleRequest struct {
	Name          string   `json:"name" binding:"required"`
	PermissionIDs []string `json:"permission_ids"`
}

func NewRoleHandler(roleService services.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// ListRoles godoc
// @Summary List roles
// @Tags Roles
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search by name"
// @Success 200 {object} map[string]interface{}
// @Router /api/roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	roles, total, err := h.roleService.ListRoles(search, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      roles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"search":    search,
	})
}

// CreateRole godoc
// @Summary Create a role
// @Tags Roles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body CreateRoleRequest true "create role"
// @Success 201 {object} services.RoleDTO
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	role, err := h.roleService.CreateRole(req.Name, req.PermissionIDs)
	if err != nil {
		if services.IsConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "role name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create role"})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// UpdateRole godoc
// @Summary Update a role
// @Tags Roles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "role id"
// @Param data body UpdateRoleRequest true "update role"
// @Success 200 {object} services.RoleDTO
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	role, err := h.roleService.UpdateRole(id, req.Name, req.PermissionIDs)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
			return
		}
		if services.IsConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "role name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update role"})
		return
	}

	c.JSON(http.StatusOK, role)
}

// DeleteRole godoc
// @Summary Soft delete a role
// @Tags Roles
// @Security BearerAuth
// @Param id path string true "role id"
// @Success 204 "deleted"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	err := h.roleService.DeleteRole(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete role"})
		return
	}
	c.Status(http.StatusNoContent)
}
