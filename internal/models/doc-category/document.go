package doc_category

import (
	enum "go-api/internal/enums"
	"time"
)

type Document struct {
	ID             string      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	DocName        string      `gorm:"column:doc_name;not null"`
	DocDescription string      `gorm:"column:doc_description"`
	CategoryID     string      `gorm:"column:category_id;type:uuid;not null;index"`
	Category       DocCategory `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	FileURL        string      `gorm:"column:file_url;not null"`
	FileType       string      `gorm:"column:file_type"`
	FileSize       int64       `gorm:"column:file_size"`

	// ✅ Business status
	Status enum.RecordStatus `gorm:"type:varchar(20);default:'active';not null"`

	// ✅ RAG processing status
	ProcessingStatus enum.ProcessingStatus `gorm:"column:processing_status;type:varchar(20);default:'pending';not null"`
	CreatedAt        time.Time             `gorm:"autoCreateTime"`
	UpdatedAt        time.Time             `gorm:"autoUpdateTime"`
}
