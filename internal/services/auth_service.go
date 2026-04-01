package services

import (
	"errors"
	"strings"
	"time"

	"go-api/internal/models"
	"go-api/internal/repositories"
	"go-api/internal/utils"
)

var (
	ErrEmailInUse        = errors.New("email already registered")
	ErrInvalidCredential = errors.New("invalid email or password")
	ErrInvalidInput      = errors.New("invalid input")
)

type AuthService interface {
	Register(email, password string) (*models.User, error)
	Login(email, password string) (string, error)
}

type authService struct {
	userRepo repositories.UserRepository
}

func NewAuthService(userRepo repositories.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Register(email, password string) (*models.User, error) {
	if !isValidEmail(email) || len(password) < 8 {
		return nil, ErrInvalidInput
	}

	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return nil, ErrEmailInUse
	} else if err != nil && err != repositories.ErrNotFound {
		return nil, err
	}

	hashed, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		PasswordHash: hashed,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if err == repositories.ErrNotFound {
			return "", ErrInvalidCredential
		}
		return "", err
	}

	if !utils.VerifyPassword(password, user.PasswordHash) {
		return "", ErrInvalidCredential
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, time.Hour)
	if err != nil {
		return "", err
	}

	return token, nil
}

func isValidEmail(email string) bool {
	return len(email) >= 3 && strings.Contains(email, "@")
}
