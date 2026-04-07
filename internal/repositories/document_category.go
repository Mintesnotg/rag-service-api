package repositories

import (
	"errors"

	"go-api/internal/enums"
	doccategory "go-api/internal/models/doc-category"

	"gorm.io/gorm"
)

var (
	ErrCatNilDB       = errors.New("nil database connection")
	ErrDocCatNotFound = errors.New("document category not found")
)

type Doc_CategoryRepository interface {
	Create(category *doccategory.DocCategory) error
	FindByID(id string) (*doccategory.DocCategory, error)
	GetAllDocCategory() ([]doccategory.DocCategory, error)
	UpdateDocCategory(category *doccategory.DocCategory) error
	DeleteDocCategory(id string) error
}

type docCategoryRepository struct {
	db *gorm.DB
}

func NewDocCategoryRepository(db *gorm.DB) Doc_CategoryRepository {
	return &docCategoryRepository{db: db}
}

func (r *docCategoryRepository) Create(category *doccategory.DocCategory) error {
	if r.db == nil {
		return ErrCatNilDB
	}
	return r.db.Create(category).Error
}

func (r *docCategoryRepository) FindByID(id string) (*doccategory.DocCategory, error) {
	if r.db == nil {
		return nil, ErrCatNilDB
	}

	var category doccategory.DocCategory
	if err := r.db.Where("id = ?", id).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocCatNotFound
		}
		return nil, err
	}
	return &category, nil
}

func (r *docCategoryRepository) GetAllDocCategory() ([]doccategory.DocCategory, error) {
	if r.db == nil {
		return nil, ErrCatNilDB
	}

	var categories []doccategory.DocCategory
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *docCategoryRepository) UpdateDocCategory(category *doccategory.DocCategory) error {
	if r.db == nil {
		return ErrCatNilDB
	}

	return r.db.Save(category).Error
}

func (r *docCategoryRepository) DeleteDocCategory(id string) error {
	if r.db == nil {
		return ErrCatNilDB
	}

	res := r.db.Model(&doccategory.DocCategory{}).
		Where("id = ?", id).
		Update("status", enums.StatusInactive)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrDocCatNotFound
	}
	return nil
}
