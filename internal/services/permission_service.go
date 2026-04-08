package services

import (
	"errors"

	"go-api/internal/repositories"

	"gorm.io/gorm"
)

type PermissionService interface {
	GetPermissionsByRoleIDs(roleIDs []string) ([]string, error)
	ListPermissions(search string, page, pageSize int) ([]PermissionDTO, int64, error)
	CreatePermission(name string) (*PermissionDTO, error)
	UpdatePermission(id, name string) (*PermissionDTO, error)
	DeletePermission(id string) error
}

type PermissionDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

type permissionService struct {
	repo repositories.PermissionRepository
}

func NewPermissionService(repo repositories.PermissionRepository) PermissionService {
	return &permissionService{repo: repo}
}

func (s *permissionService) GetPermissionsByRoleIDs(roleIDs []string) ([]string, error) {
	return s.repo.GetPermissionNamesByRoleIDs(roleIDs)
}

func (s *permissionService) ListPermissions(search string, page, pageSize int) ([]PermissionDTO, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	perms, total, err := s.repo.List(search, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	out := make([]PermissionDTO, 0, len(perms))
	for _, p := range perms {
		out = append(out, PermissionDTO{ID: p.ID, Name: p.Name, IsActive: p.IsActive})
	}
	return out, total, nil
}

func (s *permissionService) CreatePermission(name string) (*PermissionDTO, error) {
	perm, err := s.repo.Create(name)
	if err != nil {
		return nil, err
	}
	return &PermissionDTO{ID: perm.ID, Name: perm.Name, IsActive: perm.IsActive}, nil
}

func (s *permissionService) UpdatePermission(id, name string) (*PermissionDTO, error) {
	perm, err := s.repo.Update(id, name)
	if err != nil {
		return nil, err
	}
	return &PermissionDTO{ID: perm.ID, Name: perm.Name, IsActive: perm.IsActive}, nil
}

func (s *permissionService) DeletePermission(id string) error {
	err := s.repo.SoftDelete(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return err
}
