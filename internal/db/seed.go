package db

import (
	"errors"
	"log"
	"time"

	docmodels "go-api/internal/models/doc-category"
	usermodels "go-api/internal/models/user"
	"go-api/internal/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Seed(db *gorm.DB) error {
	log.Println("Starting database seeding...")

	// if err := seedRoles(db); err != nil {
	// 	return err
	// }

	// if err := seedPermissions(db); err != nil {
	// 	return err
	// }

	// if err := seedDocCategories(db); err != nil {
	// 	return err
	// }

	// if err := seedUser(db); err != nil {
	// 	return err
	// }

	// if err := SeedRolePermissions(db); err != nil {
	// 	return err
	// }

	log.Println("Database seeding completed successfully")
	return nil
}

func seedPermissions(db *gorm.DB) error {
	names := []string{
		"view_account_management",
		"view_users",
		"view_roles",
		"view_permissions",
		"view_doc_management",
		"view_hr_docs",
		"view_requirement_doc",
		"view_benefit_docs",
		"view_time_docs",
		"view_it_docs",
		"view_access_docs",
		"view_apps_docs",
		"view_security_docs",
	}

	for _, name := range names {
		var existing usermodels.Permission
		err := db.Where("name = ?", name).First(&existing).Error
		if err == nil {
			continue
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := db.Create(&usermodels.Permission{Name: name}).Error; err != nil {
			return err
		}
	}

	log.Println("Permissions seeded")
	return nil
}

func seedDocCategories(db *gorm.DB) error {
	categories := []docmodels.DocCategory{
		{
			Name:        "HR",
			Description: "Human resources documents and policies",
		},
		{
			Name:        "IT",
			Description: "Information technology documents and procedures",
		},
		{
			Name:        "Finance",
			Description: "Finance documents, reports, and controls",
		},
	}

	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "name"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"description": gorm.Expr("EXCLUDED.description"),
			"updated_at":  gorm.Expr("NOW()"),
		}),
	}).Create(&categories).Error; err != nil {
		return err
	}

	log.Println("Document categories seeded")
	return nil
}

func seedRoles(db *gorm.DB) error {
	roles := []usermodels.Role{
		{Name: "admin"},
		{Name: "user"},
	}

	for _, role := range roles {
		var existing usermodels.Role

		err := db.Where("name = ?", role.Name).First(&existing).Error
		if err == nil {
			continue // already exists
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := db.Create(&role).Error; err != nil {
			return err
		}
	}

	log.Println("Roles seeded")
	return nil
}

func seedUser(db *gorm.DB) error {
	var count int64
	if err := db.Model(&usermodels.User{}).Where("email = ?", "admin@example.com").Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	hashedPassword, err := utils.HashPassword("admin@1234")

	if err != nil {
		return err
	}

	var adminRole usermodels.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			adminRole = usermodels.Role{Name: "admin"}
			if err := db.Create(&adminRole).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	adminUser := usermodels.User{
		Email:        "admin@example.com",
		PasswordHash: hashedPassword,
		IsActive:     true,
		Roles:        []usermodels.Role{adminRole},

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.Create(&adminUser).Error; err != nil {
		return err
	}

	return nil
}

func SeedRolePermissions(db *gorm.DB) error {
	var adminRole usermodels.Role
	if err := db.Where("name = ?", "admin").Preload("Permissions").First(&adminRole).Error; err != nil {
		return err
	}

	var permissions []usermodels.Permission
	if err := db.Find(&permissions).Error; err != nil {
		return err
	}

	if err := db.Model(&adminRole).Association("Permissions").Replace(permissions); err != nil {
		return err
	}

	log.Println("Role permissions seeded")
	return nil
}
