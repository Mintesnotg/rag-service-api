package models

type Permission struct {
	ID   string `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	Name string `gorm:"not null"`
}
