package db

import (
	"errors"
	"log"
	"time"

	"go-api/internal/models"
	"go-api/internal/utils"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	log.Println("Starting database seeding...")

	if err := seedRoles(db); err != nil {
		return err
	}

	if err := seedPermissions(db); err != nil {
		return err
	}

	if err := seedUser(db); err != nil {
		return err
	}

	log.Println("Database seeding completed successfully")
	return nil
}

func seedPermissions(db *gorm.DB) error {
	perms := []models.Permission{
		{Name: "read"},
		{Name: "write"},
		{Name: "delete"},
	}

	for _, p := range perms {
		var existing models.Permission

		err := db.Where("name = ?", p.Name).First(&existing).Error
		if err == nil {
			continue
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := db.Create(&p).Error; err != nil {
			return err
		}
	}

	log.Println("Permissions seeded")
	return nil
}

func seedRoles(db *gorm.DB) error {
	roles := []models.Role{
		{Name: "admin"},
		{Name: "user"},
	}

	for _, role := range roles {
		var existing models.Role

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
	if err := db.Model(&models.User{}).Where("email = ?", "admin@example.com").Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	hashedPassword, err := utils.HashPassword("admin@1234")

	if err != nil {
		return err
	}

	var adminRole models.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			adminRole = models.Role{Name: "admin"}
			if err := db.Create(&adminRole).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	user := models.User{
		Email:        "admin@example.com",
		PasswordHash: hashedPassword,
		IsActive:     true,
		Roles:        []models.Role{adminRole},

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.Create(&user).Error; err != nil {
		return err
	}

	return nil
}
