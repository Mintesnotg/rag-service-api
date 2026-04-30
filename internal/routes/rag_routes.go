package routes

import (
	"go-api/internal/handlers"
	"go-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRAGRoutes(router *gin.Engine, handler *handlers.RAGHandler) {
	group := router.Group("/api/rag")
	group.Use(middleware.RAGQueryRateLimitMiddleware())
	group.POST("/query", handler.Query)
}
