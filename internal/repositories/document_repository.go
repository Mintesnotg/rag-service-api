package repositories

import (
	"errors"

	"go-api/internal/enums"
	doccategory "go-api/internal/models/doc-category"

	"gorm.io/gorm"
)

var (
	ErrDocumentNilDB    = errors.New("nil database connection")
	ErrDocumentNotFound = errors.New("document not found")
)

// DocumentRepository defines operations for working with documents.
type DocumentRepository interface {
	Create(document *doccategory.Document) error
	FindByID(id string) (*doccategory.Document, error)
	ListByCategory(categoryName string) ([]doccategory.Document, error)
	Update(document *doccategory.Document) error
	SoftDelete(id string) error
}

type documentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) Create(document *doccategory.Document) error {
	if r.db == nil {
		return ErrDocumentNilDB
	}
	return r.db.Create(document).Error
}

func (r *documentRepository) FindByID(id string) (*doccategory.Document, error) {
	if r.db == nil {
		return nil, ErrDocumentNilDB
	}

	var document doccategory.Document
	if err := r.db.
		Preload("Category").
		Where("id = ?", id).
		Where("status = ?", enums.StatusActive).
		First(&document).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	return &document, nil
}

func (r *documentRepository) ListByCategory(categoryName string) ([]doccategory.Document, error) {
	if r.db == nil {
		return nil, ErrDocumentNilDB
	}

	var documents []doccategory.Document

	query := r.db.
		Model(&doccategory.Document{}).
		Preload("Category").
		Joins("JOIN doc_categories ON doc_categories.id = documents.category_id").
		Where("documents.status = ?", enums.StatusActive).
		Where("doc_categories.status = ?", enums.StatusActive)

	if categoryName != "" {
		query = query.Where("LOWER(doc_categories.name) = LOWER(?)", categoryName)
	}

	if err := query.
		Order("documents.created_at DESC").
		Find(&documents).Error; err != nil {
		return nil, err
	}

	return documents, nil
}

func (r *documentRepository) Update(document *doccategory.Document) error {
	if r.db == nil {
		return ErrDocumentNilDB
	}
	return r.db.Save(document).Error
}

func (r *documentRepository) SoftDelete(id string) error {
	if r.db == nil {
		return ErrDocumentNilDB
	}

	res := r.db.
		Model(&doccategory.Document{}).
		Where("id = ?", id).
		Where("status = ?", enums.StatusActive).
		Update("status", enums.StatusInactive)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrDocumentNotFound
	}

	return nil
}
