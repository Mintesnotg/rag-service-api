package routes

import (
	"go-api/internal/handlers"
	"go-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRAGRoutes(
	router *gin.Engine,
	handler *handlers.RAGHandler,
	permMiddleware gin.HandlerFunc,
	headerCheck gin.HandlerFunc,
) {
	group := router.Group("/api/rag")
	group.Use(middleware.AuthMiddleware())
	if permMiddleware != nil {
		group.Use(permMiddleware)
	}
	group.Use(middleware.RAGQueryRateLimitMiddleware())
	if headerCheck != nil {
		group.Use(headerCheck)
	}
	group.Use(middleware.RequirePermission("query_rag"))
	group.POST("/query", handler.Query)

	sourceGroup := router.Group("/api/rag/source")
	sourceGroup.Use(middleware.AuthMiddleware())
	if permMiddleware != nil {
		sourceGroup.Use(permMiddleware)
	}
	if headerCheck != nil {
		sourceGroup.Use(headerCheck)
	}
	sourceGroup.Use(middleware.RequirePermission("download_reference_docs"))
	sourceGroup.GET("/:id", handler.GetSourceURL)
}
