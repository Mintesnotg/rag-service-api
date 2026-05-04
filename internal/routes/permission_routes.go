package routes

import (
	"go-api/internal/handlers"
	"go-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterPermissionRoutes(router *gin.Engine, handler *handlers.PermissionHandler, permMiddleware gin.HandlerFunc, headerCheck gin.HandlerFunc) {
	group := router.Group("/api/permissions")
	group.Use(middleware.AuthMiddleware())
	if permMiddleware != nil {
		group.Use(permMiddleware)
	}
	if headerCheck != nil {
		group.Use(headerCheck)
	}

	group.GET("", handler.ListPermissions)
	group.POST("", handler.CreatePermission)
	group.PUT("/:id", handler.UpdatePermission)
	group.DELETE("/:id", handler.DeletePermission)

	resolveGroup := router.Group("/api/permissions")
	resolveGroup.Use(middleware.AuthMiddleware())
	if permMiddleware != nil {
		resolveGroup.Use(permMiddleware)
	}
	resolveGroup.POST("/resolve", handler.ResolvePermissions)
}
