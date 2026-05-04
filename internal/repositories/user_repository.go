package repositories

import (
	"errors"

	"go-api/internal/models/user"

	"gorm.io/gorm"
)

var (
	ErrNotFound     = errors.New("record not found")
	ErrNilDB        = errors.New("nil database connection")
	ErrUserNotFound = errors.New("user not found")
	ErrRoleNotFound = errors.New("role not found")
)

type UserRepository interface {
	Create(user *user.User) error
	List(search string, limit, offset int) ([]user.User, int64, error)
	Update(user *user.User) error
	SoftDelete(id string) error
	FindByEmail(email string) (*user.User, error)
	FindByID(id string) (*user.User, error)
	FindRoleByID(id string) (*user.Role, error)
	FindRolesByIDs(ids []string) ([]user.Role, error)
	AssignRoleToUser(userID, roleID string) error
	ReplaceUserRoles(userID string, roleIDs []string) error
	RemoveRoleFromUser(userID, roleID string) error
	GetRolesByUserID(userID string) ([]user.Role, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *user.User) error {
	if r.db == nil {
		return ErrNilDB
	}
	return r.db.Create(user).Error
}

func (r *userRepository) List(search string, limit, offset int) ([]user.User, int64, error) {
	if r.db == nil {
		return nil, 0, ErrNilDB
	}

	var users []user.User
	var total int64

	query := r.db.Model(&user.User{}).Where("is_active = TRUE")
	if search != "" {
		query = query.Where("LOWER(email) LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("Roles", "is_active = TRUE").
		Order("email ASC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) Update(user *user.User) error {
	if r.db == nil {
		return ErrNilDB
	}
	return r.db.Save(user).Error
}

func (r *userRepository) SoftDelete(id string) error {
	if r.db == nil {
		return ErrNilDB
	}

	res := r.db.Model(&user.User{}).Where("id = ? AND is_active = TRUE", id).Update("is_active", false)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *userRepository) FindByEmail(email string) (*user.User, error) {
	if r.db == nil {
		return nil, ErrNilDB
	}

	var user user.User

	if err := r.db.Where("email = ?", email).Preload("Roles").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) FindByID(id string) (*user.User, error) {
	if r.db == nil {
		return nil, ErrNilDB
	}

	var user user.User
	if err := r.db.Where("id = ? AND is_active = TRUE", id).Preload("Roles", "is_active = TRUE").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindRoleByID(id string) (*user.Role, error) {
	if r.db == nil {
		return nil, ErrNilDB
	}

	var role user.Role
	if err := r.db.Where("id = ? AND is_active = TRUE", id).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *userRepository) FindRolesByIDs(ids []string) ([]user.Role, error) {
	if r.db == nil {
		return nil, ErrNilDB
	}
	if len(ids) == 0 {
		return []user.Role{}, nil
	}

	var roles []user.Role
	if err := r.db.Where("id IN ? AND is_active = TRUE", ids).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *userRepository) AssignRoleToUser(userID, roleID string) error {
	if r.db == nil {
		return ErrNilDB
	}

	var count int64
	if err := r.db.Table("user_roles").Where("user_id = ? AND role_id = ?", userID, roleID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// Insert association
	if err := r.db.Table("user_roles").Create(map[string]interface{}{"user_id": userID, "role_id": roleID}).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) ReplaceUserRoles(userID string, roleIDs []string) error {
	if r.db == nil {
		return ErrNilDB
	}

	u, err := r.FindByID(userID)
	if err != nil {
		return err
	}

	roles, err := r.FindRolesByIDs(roleIDs)
	if err != nil {
		return err
	}

	return r.db.Model(u).Association("Roles").Replace(roles)
}

func (r *userRepository) RemoveRoleFromUser(userID, roleID string) error {
	if r.db == nil {
		return ErrNilDB
	}

	u, err := r.FindByID(userID)
	if err != nil {
		return err
	}

	role, err := r.FindRoleByID(roleID)
	if err != nil {
		return err
	}

	return r.db.Model(u).Association("Roles").Delete(role)
}

func (r *userRepository) GetRolesByUserID(userID string) ([]user.Role, error) {
	if r.db == nil {
		return nil, ErrNilDB
	}

	var roles []user.Role
	if err := r.db.Joins("JOIN user_roles ur ON ur.role_id = roles.id").Where("ur.user_id = ?", userID).Find(&roles).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []user.Role{}, nil
		}
		return nil, err
	}
	return roles, nil
}
