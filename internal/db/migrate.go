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
	if err := ensureRAGIndexes(db); err != nil {
		return err
	}
	log.Println(" finished Migrating to database ...")

	return nil
}

func ensurePostgresExtensions(db *gorm.DB) error {
	statements := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`CREATE EXTENSION IF NOT EXISTS "vector"`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("enable postgres extension: %w", err)
		}
	}

	return nil

}

func ensureRAGIndexes(db *gorm.DB) error {
	statements := []string{
		`ALTER TABLE IF EXISTS rag_chunks ADD COLUMN IF NOT EXISTS embedding vector(768)`,
		`CREATE INDEX IF NOT EXISTS idx_rag_chunks_document_chunk ON rag_chunks(document_id, chunk_index)`,
		`CREATE INDEX IF NOT EXISTS idx_rag_chunks_embedding ON rag_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("ensure rag schema: %w", err)
		}
	}

	return nil
}
