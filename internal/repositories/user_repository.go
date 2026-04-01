package repositories

import (
	"errors"

	"go-api/internal/models"

	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("record not found")
	ErrNilDB    = errors.New("nil database connection")
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	if r.db == nil {
		return ErrNilDB
	}
	return r.db.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	if r.db == nil {
		return nil, ErrNilDB
	}

	var user models.User

	if err := r.db.Where("email = ?", email).Preload("Roles").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}
