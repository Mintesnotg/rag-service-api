package services

import (
	"errors"
	"strings"
	"time"

	models "go-api/internal/models/user"
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
	Login(email, password string) (string, []string, error)
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
	} else if err != repositories.ErrNotFound {
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

func (s *authService) Login(email, password string) (string, []string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if err == repositories.ErrNotFound {
			return "", nil, ErrInvalidCredential
		}
		return "", nil, err
	}

	if !utils.VerifyPassword(password, user.PasswordHash) {
		return "", nil, ErrInvalidCredential
	}

	roles, err := s.userRepo.GetRolesByUserID(user.ID)
	if err != nil {
		return "", nil, err
	}

	roleIDs := roleIDsFromRoles(roles)

	token, err := utils.GenerateJWT(user.ID, user.Email, roleIDs, time.Hour)
	if err != nil {
		return "", nil, err
	}

	return token, roleIDs, nil
}

func isValidEmail(email string) bool {
	return len(email) >= 3 && strings.Contains(email, "@")
}

func roleIDsFromRoles(roles []models.Role) []string {
	ids := make([]string, 0, len(roles))
	for _, r := range roles {
		ids = append(ids, r.ID)
	}
	return ids
}
