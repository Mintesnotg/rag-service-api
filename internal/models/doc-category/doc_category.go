package doc_category

import (
	enum "go-api/internal/enums"
	"time"
)

type DocCategory struct {
	ID          string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string `gorm:"not null;uniqueIndex"`
	Description string
	Status      enum.RecordStatus `gorm:"type:varchar(20);default:'active';not null"`

	Documents []Document `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}
