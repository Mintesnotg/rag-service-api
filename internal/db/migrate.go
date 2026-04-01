package db

import (
	"go-api/internal/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {

	err := db.AutoMigrate(models.MigrateModels...)

	if err != nil {
		panic("Failed to migrate models")

	}
	return nil

}
