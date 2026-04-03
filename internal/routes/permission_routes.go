package routes

import (
	"go-api/internal/handlers"
	"go-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterPermissionRoutes(router *gin.Engine, handler *handlers.PermissionHandler, permMiddleware gin.HandlerFunc) {
	group := router.Group("/api/permissions")
	group.Use(middleware.AuthMiddleware())
	if permMiddleware != nil {
		group.Use(permMiddleware)
	}
	group.POST("/", handler.GetPermissions)
}
