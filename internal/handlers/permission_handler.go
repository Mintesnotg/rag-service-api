package handlers

import (
	"net/http"

	"go-api/internal/services"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permissionService services.PermissionService
}

type PermissionRequest struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}

func NewPermissionHandler(permissionService services.PermissionService) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService}
}

// GetPermissions godoc
// @Summary Get permissions for the given role IDs

// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body PermissionRequest true "role ids"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/permissions [post]
func (h *PermissionHandler) GetPermissions(c *gin.Context) {
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
