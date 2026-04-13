package routes

import (
	"go-api/internal/handlers"
	"go-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(router *gin.Engine, roleHandler *handlers.RoleHandler, permMiddleware gin.HandlerFunc, headerCheck gin.HandlerFunc) {
	rolesGroup := router.Group("/api/roles")
	rolesGroup.Use(middleware.AuthMiddleware())
	if permMiddleware != nil {
		rolesGroup.Use(permMiddleware)
	}
	if headerCheck != nil {
		rolesGroup.Use(headerCheck)
	}

	rolesGroup.GET("", roleHandler.ListRoles)
	rolesGroup.POST("", roleHandler.CreateRole)
	rolesGroup.PUT(":id", roleHandler.UpdateRole)
	rolesGroup.DELETE(":id", roleHandler.DeleteRole)
}
