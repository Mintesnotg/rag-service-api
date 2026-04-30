package routes

import (
	"go-api/internal/handlers"
	"go-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(
	router *gin.Engine,
	handler *handlers.UserHandler,
	permMiddleware gin.HandlerFunc,
	headerCheck gin.HandlerFunc,
) {
	group := router.Group("/api/users")
	group.Use(middleware.AuthMiddleware())
	if permMiddleware != nil {
		group.Use(permMiddleware)
	}
	if headerCheck != nil {
		group.Use(headerCheck)
	}

	group.GET("", handler.ListUsers)
	group.POST("", handler.CreateUser)
	group.PUT("/:id", handler.UpdateUser)
	group.DELETE("/:id", handler.DeleteUser)

	group.GET("/:id/roles", handler.GetUserRoles)
	group.POST("/:id/roles", handler.AssignUserRoles)
	group.DELETE("/:id/roles/:roleId", handler.RemoveUserRole)
}
