package routes

import (
	"go-api/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.Engine, authHandler *handlers.AuthHandler) {
	auth := router.Group("/api/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
}
