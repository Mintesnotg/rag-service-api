package middleware

import (
	"net/http"

	"go-api/internal/services"

	"github.com/gin-gonic/gin"
)

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

// RequirePermission ensures the hydrated permissions contain the required one.
func RequirePermission(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get("permissions")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "unauthorized: permissions not loaded",
			})
			return
		}

		perms, ok := value.([]string)
		if !ok || len(perms) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "unauthorized: no permissions assigned",
			})
			return
		}

		for _, p := range perms {
			if p == required {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "unauthorized: missing required permission",
		})
	}
}
