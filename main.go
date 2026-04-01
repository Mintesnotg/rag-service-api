package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"go-api/internal/db"
	"go-api/internal/handlers"
	"go-api/internal/middleware"
	"go-api/internal/repositories"
	"go-api/internal/routes"
	"go-api/internal/services"
)

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

	userRepo := repositories.NewUserRepository(conn)
	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	routes.RegisterAuthRoutes(router, authHandler)

	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	protected.GET("/protected", func(c *gin.Context) {
		email, _ := c.Get("userEmail")
		c.JSON(200, gin.H{"message": "authorized", "email": email})
	})

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
