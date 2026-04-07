package services

import (
	"errors"
	"strings"

	doccategory "go-api/internal/models/doc-category"
	"go-api/internal/repositories"
)

var (
	ErrDocCategoryInvalidInput = errors.New("invalid document category input")
	ErrDocCategoryNotFound     = errors.New("document category not found")
)

type DocCategoryService interface {
	CreateDocCategory(name, description string) (*doccategory.DocCategory, error)
	UpdateDocCategory(id, name, description string) (*doccategory.DocCategory, error)
	DeleteDocCategory(id string) error
	GetAllDocCategory() ([]doccategory.DocCategory, error)
	GetDocCategoryByID(id string) (*doccategory.DocCategory, error)
}

type docCategoryService struct {
	repo repositories.Doc_CategoryRepository
}

func NewDocCategoryService(repo repositories.Doc_CategoryRepository) DocCategoryService {
	return &docCategoryService{repo: repo}
}

func (s *docCategoryService) CreateDocCategory(name, description string) (*doccategory.DocCategory, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" {
		return nil, ErrDocCategoryInvalidInput
	}

	category := &doccategory.DocCategory{
		Name:        name,
		Description: description,
	}

	if err := s.repo.Create(category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *docCategoryService) UpdateDocCategory(id, name, description string) (*doccategory.DocCategory, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if id == "" || name == "" {
		return nil, ErrDocCategoryInvalidInput
	}

	category, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrDocCatNotFound) {
			return nil, ErrDocCategoryNotFound
		}
		return nil, err
	}

	category.Name = name
	category.Description = description

	if err := s.repo.UpdateDocCategory(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *docCategoryService) DeleteDocCategory(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrDocCategoryInvalidInput
	}

	if err := s.repo.DeleteDocCategory(id); err != nil {
		if errors.Is(err, repositories.ErrDocCatNotFound) {
			return ErrDocCategoryNotFound
		}
		return err
	}
	return nil
}

func (s *docCategoryService) GetAllDocCategory() ([]doccategory.DocCategory, error) {
	return s.repo.GetAllDocCategory()
}

func (s *docCategoryService) GetDocCategoryByID(id string) (*doccategory.DocCategory, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrDocCategoryInvalidInput
	}

	category, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrDocCatNotFound) {
			return nil, ErrDocCategoryNotFound
		}
		return nil, err
	}

	return category, nil
}
