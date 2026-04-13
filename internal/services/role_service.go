package services

import (
	"errors"

	"go-api/internal/models/user"
	"go-api/internal/repositories"

	"gorm.io/gorm"
)

type RoleService interface {
	ListRoles(search string, page, pageSize int) ([]RoleDTO, int64, error)
	CreateRole(name string, permissionIDs []string) (*RoleDTO, error)
	UpdateRole(id, name string, permissionIDs []string) (*RoleDTO, error)
	DeleteRole(id string) error
}

type RoleDTO struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	IsActive        bool            `json:"is_active"`
	Permissions     []PermissionDTO `json:"permissions"`
	PermissionCount int             `json:"permission_count"`
}

type roleService struct {
	repo repositories.RoleRepository
}

func NewRoleService(repo repositories.RoleRepository) RoleService {
	return &roleService{repo: repo}
}

func (s *roleService) ListRoles(search string, page, pageSize int) ([]RoleDTO, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	roles, total, err := s.repo.List(search, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	out := make([]RoleDTO, 0, len(roles))
	for _, r := range roles {
		out = append(out, toRoleDTO(r))
	}
	return out, total, nil
}

func (s *roleService) CreateRole(name string, permissionIDs []string) (*RoleDTO, error) {
	role, err := s.repo.Create(name, permissionIDs)
	if err != nil {
		return nil, err
	}
	dto := toRoleDTO(*role)
	return &dto, nil
}

func (s *roleService) UpdateRole(id, name string, permissionIDs []string) (*RoleDTO, error) {
	role, err := s.repo.Update(id, name, permissionIDs)
	if err != nil {
		return nil, err
	}
	dto := toRoleDTO(*role)
	return &dto, nil
}

func (s *roleService) DeleteRole(id string) error {
	err := s.repo.SoftDelete(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return err
}

func toRoleDTO(r user.Role) RoleDTO {
	dto := RoleDTO{
		ID:       r.ID,
		Name:     r.Name,
		IsActive: r.IsActive,
	}
	for _, p := range r.Permissions {
		dto.Permissions = append(dto.Permissions, PermissionDTO{ID: p.ID, Name: p.Name, IsActive: p.IsActive})
	}
	dto.PermissionCount = len(dto.Permissions)
	return dto
}
