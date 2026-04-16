package routes

import (
	"go-api/internal/handlers"
	"go-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterDocumentRoutes(
	router *gin.Engine,
	handler *handlers.DocumentHandler,
	permMiddleware gin.HandlerFunc,
	headerCheck gin.HandlerFunc,
) {
	group := router.Group("/api/documents")
	group.Use(middleware.AuthMiddleware())
	if permMiddleware != nil {
		group.Use(permMiddleware)
	}
	group.Use(middleware.RequirePermission("document.manage"))
	if headerCheck != nil {
		group.Use(headerCheck)
	}

	group.GET("", handler.ListDocuments)
	group.POST("", handler.CreateDocument)
	group.PUT("/:id", handler.UpdateDocument)
	group.DELETE("/:id", handler.DeleteDocument)
	group.GET("/:id/download", handler.GetDocumentDownloadURL)
}
