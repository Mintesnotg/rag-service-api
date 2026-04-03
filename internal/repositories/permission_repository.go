package repositories

import (
	// "go-api/internal/models"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	GetPermissionNamesByRoleIDs(roleIDs []string) ([]string, error)
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) GetPermissionNamesByRoleIDs(roleIDs []string) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}

	var names []string
	err := r.db.
		Table("permissions p").
		Select("DISTINCT p.name").
		Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
		Where("rp.role_id IN ?", roleIDs).
		Scan(&names).Error

	if err != nil {
		return nil, err
	}

	return names, nil
}
