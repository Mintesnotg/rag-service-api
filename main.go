package main

import (
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go-api/internal/db"
	"go-api/internal/handlers"
	"go-api/internal/middleware"
	"go-api/internal/repositories"
	"go-api/internal/routes"
	"go-api/internal/services"
	"go-api/internal/storage"

	_ "go-api/docs"
)

// @title Smart Doc API
// @version 1.0
// @description API documentation for Smart Doc service.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	conn, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	if err := db.Migrate(conn); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	if err := db.Seed(conn); err != nil {
		log.Fatalf("seeding failed: %v", err)
	}

	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	userRepo := repositories.NewUserRepository(conn)
	roleRepo := repositories.NewRoleRepository(conn)
	permissionRepo := repositories.NewPermissionRepository(conn)
	docCategoryRepo := repositories.NewDocCategoryRepository(conn)
	documentRepo := repositories.NewDocumentRepository(conn)

	minioStorage, err := storage.NewMinIOStorage()
	if err != nil {
		log.Fatalf("minio initialization failed: %v", err)
	}

	authService := services.NewAuthService(userRepo)
	roleService := services.NewRoleService(roleRepo)
	permissionService := services.NewPermissionService(permissionRepo)
	docCategoryService := services.NewDocCategoryService(docCategoryRepo)
	documentService := services.NewDocumentService(documentRepo, docCategoryRepo, minioStorage)

	authHandler := handlers.NewAuthHandler(authService)
	roleHandler := handlers.NewRoleHandler(roleService)
	permissionHandler := handlers.NewPermissionHandler(permissionService)
	docCategoryHandler := handlers.NewDocCategoryHandler(docCategoryService)
	documentHandler := handlers.NewDocumentHandler(documentService)

	permissionHydrator := middleware.PermissionsMiddleware(permissionService)
	headerPermissionCheck := middleware.RequireHeaderPermission()

	routes.RegisterAuthRoutes(router, authHandler)
	routes.RegisterPermissionRoutes(router, permissionHandler, permissionHydrator, headerPermissionCheck)
	routes.RegisterRoleRoutes(router, roleHandler, permissionHydrator, headerPermissionCheck)
	routes.RegisterDocCategoryRoutes(router, docCategoryHandler, permissionHydrator)
	routes.RegisterDocumentRoutes(router, documentHandler, permissionHydrator, headerPermissionCheck)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
