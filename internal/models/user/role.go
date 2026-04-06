package user

type Role struct {
	ID string `gorm:"type:uuid;default:gen_random_uuid();primary_key"`

	Name string `gorm:"not null"`

	// Permissions []Permission `grom:"many2many:role_permissions"`

	Permissions []Permission `gorm:"many2many:role_permissions"`
}
