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
	group.Use(middleware.RequirePermission("document.manage"))
	if headerCheck != nil {
		group.Use(headerCheck)
	}

	group.POST("/query", handler.Query)
}
