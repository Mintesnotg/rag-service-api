package db

import (
	"fmt"
	models "go-api/internal/models"
	"log"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {

	log.Println("Migrating to database ...")
	if err := ensurePostgresExtensions(db); err != nil {
		return err
	}

	if err := db.AutoMigrate(models.MigrateModels...); err != nil {
		return fmt.Errorf("auto migrate models: %w", err)
	}
	log.Println(" finished Migrating to database ...")

	return nil
}

func ensurePostgresExtensions(db *gorm.DB) error {
	statements := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("enable postgres extension: %w", err)
		}
	}

	return nil

}
