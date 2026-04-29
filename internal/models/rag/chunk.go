package rag

import "time"

type Chunk struct {
	ID         string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	DocumentID string    `gorm:"column:document_id;type:uuid;not null;index"`
	ChunkIndex int       `gorm:"column:chunk_index;not null"`
	Content    string    `gorm:"column:content;type:text;not null"`
	Metadata   string    `gorm:"column:metadata;type:jsonb;default:'{}';not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (Chunk) TableName() string {
	return "rag_chunks"
}
