package repositories

import (
	"errors"

	user "go-api/internal/models/user"

	"gorm.io/gorm"
)

type RoleRepository interface {
	List(search string, limit, offset int) ([]user.Role, int64, error)
	Create(name string, permissionIDs []string) (*user.Role, error)
	Update(id, name string, permissionIDs []string) (*user.Role, error)
	SoftDelete(id string) error
	GetByID(id string) (*user.Role, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) List(search string, limit, offset int) ([]user.Role, int64, error) {
	var roles []user.Role
	var total int64

	query := r.db.Model(&user.Role{}).Where("is_active = TRUE")
	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Permissions", "is_active = TRUE").Order("name ASC").Limit(limit).Offset(offset).Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (r *roleRepository) Create(name string, permissionIDs []string) (*user.Role, error) {
	role := user.Role{Name: name, IsActive: true}
	tx := r.db.Begin()
	if err := tx.Create(&role).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if len(permissionIDs) > 0 {
		var perms []user.Permission
		if err := tx.Where("id IN ? AND is_active = TRUE", permissionIDs).Find(&perms).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := tx.Model(&role).Association("Permissions").Replace(perms); err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	if err := r.db.Preload("Permissions", "is_active = TRUE").First(&role, "id = ?", role.ID).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Update(id, name string, permissionIDs []string) (*user.Role, error) {
	tx := r.db.Begin()
	var role user.Role
	if err := tx.Preload("Permissions", "is_active = TRUE").First(&role, "id = ? AND is_active = TRUE", id).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	role.Name = name
	if err := tx.Save(&role).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var perms []user.Permission
	if len(permissionIDs) > 0 {
		if err := tx.Where("id IN ? AND is_active = TRUE", permissionIDs).Find(&perms).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Model(&role).Association("Permissions").Replace(perms); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if err := r.db.Preload("Permissions", "is_active = TRUE").First(&role, "id = ?", role.ID).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *roleRepository) SoftDelete(id string) error {
	res := r.db.Model(&user.Role{}).Where("id = ?", id).Update("is_active", false)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *roleRepository) GetByID(id string) (*user.Role, error) {
	var role user.Role
	if err := r.db.Preload("Permissions", "is_active = TRUE").First(&role, "id = ? AND is_active = TRUE", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &role, nil
}
