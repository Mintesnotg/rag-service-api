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
	ErrUserInvalidInput = errors.New("invalid user input")
)

type UserService interface {
	ListUsers(search string, page, pageSize int) ([]UserDTO, int64, error)
	CreateUser(email, password string, roleIDs []string) (*UserDTO, error)
	UpdateUser(id, email, password string, roleIDs []string) (*UserDTO, error)
	DeleteUser(id string) error
	GetUserRoles(userID string) ([]RoleLiteDTO, error)
	AssignRoles(userID string, roleIDs []string) ([]RoleLiteDTO, error)
	RemoveRole(userID, roleID string) ([]RoleLiteDTO, error)
}

type UserDTO struct {
	ID        string        `json:"id"`
	Email     string        `json:"email"`
	IsActive  bool          `json:"is_active"`
	RoleIDs   []string      `json:"role_ids"`
	Roles     []RoleLiteDTO `json:"roles"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type RoleLiteDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type userService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) ListUsers(search string, page, pageSize int) ([]UserDTO, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	users, total, err := s.repo.List(strings.TrimSpace(search), pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	out := make([]UserDTO, 0, len(users))
	for _, u := range users {
		out = append(out, toUserDTO(u))
	}
	return out, total, nil
}

func (s *userService) CreateUser(email, password string, roleIDs []string) (*UserDTO, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if !isValidEmail(email) || len(password) < 8 {
		return nil, ErrUserInvalidInput
	}

	if _, err := s.repo.FindByEmail(email); err == nil {
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

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	if err := s.repo.ReplaceUserRoles(user.ID, roleIDs); err != nil {
		return nil, err
	}

	created, err := s.repo.FindByID(user.ID)
	if err != nil {
		return nil, err
	}
	dto := toUserDTO(*created)
	return &dto, nil
}

func (s *userService) UpdateUser(id, email, password string, roleIDs []string) (*UserDTO, error) {
	id = strings.TrimSpace(id)
	email = strings.TrimSpace(strings.ToLower(email))
	if id == "" || !isValidEmail(email) {
		return nil, ErrUserInvalidInput
	}

	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(existing.Email, email) {
		if _, err := s.repo.FindByEmail(email); err == nil {
			return nil, ErrEmailInUse
		} else if err != repositories.ErrNotFound {
			return nil, err
		}
	}

	existing.Email = email
	if strings.TrimSpace(password) != "" {
		if len(password) < 8 {
			return nil, ErrUserInvalidInput
		}
		hashed, err := utils.HashPassword(password)
		if err != nil {
			return nil, err
		}
		existing.PasswordHash = hashed
	}

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}

	if err := s.repo.ReplaceUserRoles(existing.ID, roleIDs); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(existing.ID)
	if err != nil {
		return nil, err
	}
	dto := toUserDTO(*updated)
	return &dto, nil
}

func (s *userService) DeleteUser(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrUserInvalidInput
	}
	return s.repo.SoftDelete(id)
}

func (s *userService) GetUserRoles(userID string) ([]RoleLiteDTO, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUserInvalidInput
	}
	roles, err := s.repo.GetRolesByUserID(userID)
	if err != nil {
		return nil, err
	}
	return toRoleLites(roles), nil
}

func (s *userService) AssignRoles(userID string, roleIDs []string) ([]RoleLiteDTO, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUserInvalidInput
	}
	if err := s.repo.ReplaceUserRoles(userID, roleIDs); err != nil {
		return nil, err
	}
	roles, err := s.repo.GetRolesByUserID(userID)
	if err != nil {
		return nil, err
	}
	return toRoleLites(roles), nil
}

func (s *userService) RemoveRole(userID, roleID string) ([]RoleLiteDTO, error) {
	userID = strings.TrimSpace(userID)
	roleID = strings.TrimSpace(roleID)
	if userID == "" || roleID == "" {
		return nil, ErrUserInvalidInput
	}
	if err := s.repo.RemoveRoleFromUser(userID, roleID); err != nil {
		return nil, err
	}
	roles, err := s.repo.GetRolesByUserID(userID)
	if err != nil {
		return nil, err
	}
	return toRoleLites(roles), nil
}

func toUserDTO(u models.User) UserDTO {
	roleIDs := make([]string, 0, len(u.Roles))
	roleLites := make([]RoleLiteDTO, 0, len(u.Roles))
	for _, r := range u.Roles {
		roleIDs = append(roleIDs, r.ID)
		roleLites = append(roleLites, RoleLiteDTO{ID: r.ID, Name: r.Name})
	}

	return UserDTO{
		ID:        u.ID,
		Email:     u.Email,
		IsActive:  u.IsActive,
		RoleIDs:   roleIDs,
		Roles:     roleLites,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toRoleLites(roles []models.Role) []RoleLiteDTO {
	out := make([]RoleLiteDTO, 0, len(roles))
	for _, r := range roles {
		out = append(out, RoleLiteDTO{ID: r.ID, Name: r.Name})
	}
	return out
}
