package routes

import (
	"go-api/internal/handlers"
	"go-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterDocCategoryRoutes(router *gin.Engine, handler *handlers.DocCategoryHandler, permMiddleware gin.HandlerFunc) {
	group := router.Group("/api/doc-categories")
	group.Use(middleware.AuthMiddleware())
	if permMiddleware != nil {
		group.Use(permMiddleware)
	}

	group.POST("", handler.CreateDocCategory)
	group.GET("", handler.GetAllDocCategories)
	group.GET("/:id", handler.GetDocCategoryByID)
	group.PUT("/:id", handler.UpdateDocCategory)
	group.DELETE("/:id", handler.DeleteDocCategory)
}

