package services

import "go-api/internal/repositories"

type PermissionService interface {
	GetPermissionsByRoleIDs(roleIDs []string) ([]string, error)
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
