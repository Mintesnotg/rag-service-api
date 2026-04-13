package user

type Role struct {
	ID          string       `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	Name        string       `gorm:"not null;uniqueIndex"`
	IsActive    bool         `gorm:"not null;default:true"`
	Permissions []Permission `gorm:"many2many:role_permissions"`
}
