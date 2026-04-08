package middleware

import (
	"net/http"
	"strings"

	"go-api/internal/services"

	"github.com/gin-gonic/gin"
)

const permissionHeader = "X-Permission"

// PermissionsMiddleware hydrates the request with permissions resolved from role IDs in the JWT.
func PermissionsMiddleware(permissionService services.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get("role_ids")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "unauthorized: missing role assignments",
			})
			return
		}

		roleIDs, ok := value.([]string)
		if !ok || len(roleIDs) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "unauthorized: no roles found for user",
			})
			return
		}

		perms, err := permissionService.GetPermissionsByRoleIDs(roleIDs)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to load permissions",
			})
			return
		}

		c.Set("permissions", perms)
		c.Next()
	}
}

// RequirePermission ensures the hydrated permissions contain the required one (route-level check).
func RequirePermission(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if hasPermission(c, required) {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "unauthorized: missing required permission",
		})
	}
}

// RequireHeaderPermission validates the permission claim sent by the client in the header.
// Clients should set X-Permission: <permission_name> for every protected request.
func RequireHeaderPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		requested := strings.TrimSpace(c.GetHeader(permissionHeader))
		if requested == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "unauthorized: missing permission claim header",
			})
			return
		}

		if hasPermission(c, requested) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "unauthorized: missing required permission",
		})
	}
}

func hasPermission(c *gin.Context, required string) bool {
	value, exists := c.Get("permissions")
	if !exists {
		return false
	}
	perms, ok := value.([]string)
	if !ok || len(perms) == 0 {
		return false
	}
	for _, p := range perms {
		if p == required {
			return true
		}
	}
	return false
}
