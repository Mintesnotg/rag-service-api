package main

import (
	"log"

	"go-api/internal/db"
)

func main() {
	db.ConnectDB()

	if err := db.Migrate(db.DB); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	if err := db.Seed(db.DB); err != nil {
		log.Fatalf("seeding failed: %v", err)
	}
}
