package user

import "time"

type User struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	Email        string    `gorm:"not null;uniqueindex"`
	PasswordHash string    `gorm:"not null"`
	IsActive     bool      `gorm:"default:true"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	Roles        []Role    `gorm:"many2many:user_roles"`
}
