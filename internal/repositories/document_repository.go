package repositories

import (
	"errors"

	"go-api/internal/enums"
	doccategory "go-api/internal/models/doc-category"

	"gorm.io/gorm"
)

var (
	ErrDocNilDB         = errors.New("nil database connection")
	ErrDocumentNotFound = errors.New("document not found")
)

// DocumentRepository defines operations for working with documents.
type DocumentRepository interface {
	Create(document *doccategory.Document) error
	FindDocByID(id string) (*doccategory.Document, error)
	GetAllDocument() ([]doccategory.Document, error)
	UpdateDocument(document *doccategory.Document) error
	DeleteDocument(id string) error
}

type documentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) Create(document *doccategory.Document) error {
	if r.db == nil {
		return ErrDocNilDB
	}
	return r.db.Create(document).Error
}

func (r *documentRepository) FindDocByID(id string) (*doccategory.Document, error) {
	if r.db == nil {
		return nil, ErrDocNilDB
	}

	var document doccategory.Document
	if err := r.db.Where("id = ?", id).Preload("Category").First(&document).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	return &document, nil
}

func (r *documentRepository) GetAllDocument() ([]doccategory.Document, error) {
	if r.db == nil {
		return nil, ErrDocNilDB
	}

	var documents []doccategory.Document
	if err := r.db.Preload("Category").Find(&documents).Error; err != nil {
		return nil, err
	}
	return documents, nil
}

func (r *documentRepository) UpdateDocument(document *doccategory.Document) error {
	if r.db == nil {
		return ErrDocNilDB
	}
	return r.db.Save(document).Error
}

func (r *documentRepository) DeleteDocument(id string) error {
	if r.db == nil {
		return ErrDocNilDB
	}

	res := r.db.Model(&doccategory.Document{}).
		Where("id = ?", id).
		Update("status", enums.StatusInactive)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrDocumentNotFound
	}
	return nil
}
