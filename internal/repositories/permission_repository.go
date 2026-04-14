package repositories

import (
	"errors"

	user "go-api/internal/models/user"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	GetPermissionNamesByRoleIDs(roleIDs []string) ([]string, error)
	List(search string, limit, offset int) ([]user.Permission, int64, error)
	Create(name string) (*user.Permission, error)
	Update(id, name string) (*user.Permission, error)
	SoftDelete(id string) error
	GetByID(id string) (*user.Permission, error)
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
		Where("rp.role_id IN ? AND p.is_active = TRUE", roleIDs).
		Scan(&names).Error

	if err != nil {
		return nil, err
	}

	return names, nil
}

func (r *permissionRepository) List(search string, limit, offset int) ([]user.Permission, int64, error) {
	var perms []user.Permission
	var total int64

	query := r.db.Model(&user.Permission{}).Where("is_active = TRUE")
	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// if err := query.Order("name ASC").Limit(limit).Offset(offset).Find(&perms).Error; err != nil {
	// 	return nil, 0, err
	// }

	if err := query.Order("name ASC").Find(&perms).Error; err != nil {
		return nil, 0, err
	}

	return perms, total, nil
}

func (r *permissionRepository) Create(name string) (*user.Permission, error) {
	perm := user.Permission{Name: name, IsActive: true}
	if err := r.db.Create(&perm).Error; err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *permissionRepository) Update(id, name string) (*user.Permission, error) {
	perm := user.Permission{}
	if err := r.db.First(&perm, "id = ? AND is_active = TRUE", id).Error; err != nil {
		return nil, err
	}
	perm.Name = name
	if err := r.db.Save(&perm).Error; err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *permissionRepository) SoftDelete(id string) error {
	res := r.db.Model(&user.Permission{}).Where("id = ?", id).Update("is_active", false)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *permissionRepository) GetByID(id string) (*user.Permission, error) {
	var perm user.Permission
	if err := r.db.First(&perm, "id = ? AND is_active = TRUE", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &perm, nil
}
