package routes

import (
	"go-api/internal/handlers"
	"go-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(router *gin.Engine, roleHandler *handlers.RoleHandler, permMiddleware gin.HandlerFunc) {
	rolesGroup := router.Group("/api/roles")
	rolesGroup.Use(middleware.AuthMiddleware())
	if permMiddleware != nil {
		rolesGroup.Use(permMiddleware)
	}
	rolesGroup.POST("/assign", roleHandler.AssignRole)
}
