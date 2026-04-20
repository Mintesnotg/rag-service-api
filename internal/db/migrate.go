package db

import (
	"fmt"
	models "go-api/internal/models"
	"log"
	"os"
	"strings"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {

	log.Println("Migrating to database ...")
	vectorEnabled, err := ensurePostgresExtensions(db)
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(models.MigrateModels...); err != nil {
		return fmt.Errorf("auto migrate models: %w", err)
	}
	if vectorEnabled {
		if err := ensureRAGIndexes(db); err != nil {
			return err
		}
	} else {
		log.Println("pgvector extension is unavailable; skipping vector-specific RAG schema updates")
	}
	log.Println(" finished Migrating to database ...")

	return nil
}

func ensurePostgresExtensions(db *gorm.DB) (bool, error) {
	requiredStatements := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
	}

	for _, statement := range requiredStatements {
		if err := db.Exec(statement).Error; err != nil {
			return false, fmt.Errorf("enable postgres extension: %w", err)
		}
	}

	vectorErr := db.Exec(`CREATE EXTENSION IF NOT EXISTS "vector"`).Error
	if vectorErr != nil {
		if requiresPgVector() {
			return false, fmt.Errorf("enable postgres extension: %w", vectorErr)
		}

		log.Printf("warning: pgvector extension not enabled (%v)", vectorErr)
		return false, nil
	}

	return true, nil

}

func ensureRAGIndexes(db *gorm.DB) error {
	if !db.Migrator().HasTable("rag_chunks") {
		log.Println("rag_chunks table not found; skipping RAG index setup")
		return nil
	}

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

func requiresPgVector() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("REQUIRE_PGVECTOR")))
	return value == "1" || value == "true" || value == "yes"
}
