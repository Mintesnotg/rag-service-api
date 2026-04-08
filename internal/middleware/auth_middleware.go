package middleware

import (
	"net/http"
	"strings"

	"go-api/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}

		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("role_ids", claims.RoleIDs)

		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer") {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
	}

	if cookie, err := c.Cookie("__Host-token"); err == nil && cookie != "" {
		return cookie
	}
	if cookie, err := c.Cookie("token"); err == nil && cookie != "" {
		return cookie
	}
	if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
		return cookie
	}

	return ""
}
