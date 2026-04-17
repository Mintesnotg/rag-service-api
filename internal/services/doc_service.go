package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"go-api/internal/enums"
	doccategory "go-api/internal/models/doc-category"
	"go-api/internal/repositories"
	"go-api/internal/storage"
)

var (
	ErrDocumentInvalidInput = errors.New("invalid document input")
	ErrDocumentNotFound     = errors.New("document not found")
	ErrDocumentCategory     = errors.New("document category not found")
)

type DocumentService interface {
	CreateDocument(ctx context.Context, input CreateDocumentInput) (*DocumentDTO, error)
	ListDocuments(ctx context.Context, categoryName string) ([]DocumentDTO, error)
	UpdateDocument(ctx context.Context, id string, input UpdateDocumentInput) (*DocumentDTO, error)
	DeleteDocument(ctx context.Context, id string) error
	GetDownloadURL(ctx context.Context, id string) (string, error)
}

type CreateDocumentInput struct {
	DocName        string
	DocDescription string
	CategoryName   string
	File           *multipart.FileHeader
}

type UpdateDocumentInput struct {
	DocName        string
	DocDescription string
	CategoryName   string
	File           *multipart.FileHeader
}

type DocumentDTO struct {
	ID               string    `json:"id"`
	DocName          string    `json:"doc_name"`
	DocDescription   string    `json:"doc_description"`
	CategoryID       string    `json:"category_id"`
	CategoryName     string    `json:"category_name"`
	FileURL          string    `json:"file_url"`
	FileName         string    `json:"file_name"`
	FileType         string    `json:"file_type"`
	FileSize         int64     `json:"file_size"`
	Status           string    `json:"status"`
	ProcessingStatus string    `json:"processing_status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type documentService struct {
	repo         repositories.DocumentRepository
	categoryRepo repositories.Doc_CategoryRepository
	fileStore    storage.ObjectStorage
	ragService   RAGService
}

func NewDocumentService(
	repo repositories.DocumentRepository,
	categoryRepo repositories.Doc_CategoryRepository,
	fileStore storage.ObjectStorage,
	ragService RAGService,
) DocumentService {
	return &documentService{
		repo:         repo,
		categoryRepo: categoryRepo,
		fileStore:    fileStore,
		ragService:   ragService,
	}
}

func (s *documentService) CreateDocument(ctx context.Context, input CreateDocumentInput) (*DocumentDTO, error) {
	docName := strings.TrimSpace(input.DocName)
	docDescription := strings.TrimSpace(input.DocDescription)
	categoryName := strings.TrimSpace(input.CategoryName)

	if docName == "" || categoryName == "" || input.File == nil {
		return nil, ErrDocumentInvalidInput
	}

	category, err := s.categoryRepo.FindByName(categoryName)
	if err != nil {
		if errors.Is(err, repositories.ErrDocCatNotFound) {
			return nil, ErrDocumentCategory
		}
		return nil, err
	}

	objectKey, err := s.uploadFile(ctx, category.Name, input.File)
	if err != nil {
		return nil, err
	}

	document := &doccategory.Document{
		DocName:          docName,
		DocDescription:   docDescription,
		CategoryID:       category.ID,
		FileURL:          objectKey,
		FileType:         detectContentType(input.File),
		FileSize:         input.File.Size,
		Status:           enums.StatusActive,
		ProcessingStatus: enums.ProcessingPending,
	}

	if err := s.repo.Create(document); err != nil {
		_ = s.fileStore.Delete(ctx, objectKey)
		return nil, err
	}

	document.Category = *category
	s.scheduleRAGIndexing(ctx, document)
	dto := mapDocument(document)
	return &dto, nil
}

func (s *documentService) ListDocuments(ctx context.Context, categoryName string) ([]DocumentDTO, error) {
	_ = ctx

	categoryName = strings.TrimSpace(categoryName)
	if categoryName == "" {
		return nil, ErrDocumentInvalidInput
	}

	documents, err := s.repo.ListByCategory(categoryName)
	if err != nil {
		return nil, err
	}

	result := make([]DocumentDTO, 0, len(documents))
	for i := range documents {
		result = append(result, mapDocument(&documents[i]))
	}
	return result, nil
}

func (s *documentService) UpdateDocument(ctx context.Context, id string, input UpdateDocumentInput) (*DocumentDTO, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrDocumentInvalidInput
	}

	document, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrDocumentNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	docName := strings.TrimSpace(input.DocName)
	docDescription := strings.TrimSpace(input.DocDescription)
	categoryName := strings.TrimSpace(input.CategoryName)

	if docName == "" || categoryName == "" {
		return nil, ErrDocumentInvalidInput
	}

	category, err := s.categoryRepo.FindByName(categoryName)
	if err != nil {
		if errors.Is(err, repositories.ErrDocCatNotFound) {
			return nil, ErrDocumentCategory
		}
		return nil, err
	}

	oldObjectKey := document.FileURL
	document.DocName = docName
	document.DocDescription = docDescription
	document.CategoryID = category.ID
	document.Category = *category

	if input.File != nil {
		objectKey, uploadErr := s.uploadFile(ctx, category.Name, input.File)
		if uploadErr != nil {
			return nil, uploadErr
		}
		document.FileURL = objectKey
		document.FileType = detectContentType(input.File)
		document.FileSize = input.File.Size
	}

	if err := s.repo.Update(document); err != nil {
		return nil, err
	}

	if input.File != nil && oldObjectKey != "" && oldObjectKey != document.FileURL {
		_ = s.fileStore.Delete(ctx, oldObjectKey)
	}
	if input.File != nil {
		s.scheduleRAGIndexing(ctx, document)
	}

	dto := mapDocument(document)
	return &dto, nil
}

func (s *documentService) DeleteDocument(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrDocumentInvalidInput
	}

	if err := s.repo.SoftDelete(id); err != nil {
		if errors.Is(err, repositories.ErrDocumentNotFound) {
			return ErrDocumentNotFound
		}
		return err
	}
	if s.ragService != nil {
		_ = s.ragService.RemoveDocument(ctx, id)
	}
	return nil
}

func (s *documentService) GetDownloadURL(ctx context.Context, id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", ErrDocumentInvalidInput
	}

	document, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrDocumentNotFound) {
			return "", ErrDocumentNotFound
		}
		return "", err
	}

	return s.fileStore.PresignedGetURL(ctx, document.FileURL, 15*time.Minute)
}

func (s *documentService) uploadFile(ctx context.Context, categoryName string, fileHeader *multipart.FileHeader) (string, error) {
	openedFile, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer openedFile.Close()

	objectKey := buildObjectKey(categoryName, fileHeader.Filename)
	contentType := detectContentType(fileHeader)
	if err := s.fileStore.Upload(ctx, objectKey, openedFile, fileHeader.Size, contentType); err != nil {
		return "", err
	}
	return objectKey, nil
}

func mapDocument(document *doccategory.Document) DocumentDTO {
	categoryName := ""
	if document.Category.ID != "" {
		categoryName = document.Category.Name
	}

	return DocumentDTO{
		ID:               document.ID,
		DocName:          document.DocName,
		DocDescription:   document.DocDescription,
		CategoryID:       document.CategoryID,
		CategoryName:     categoryName,
		FileURL:          document.FileURL,
		FileName:         extractFileName(document.FileURL),
		FileType:         document.FileType,
		FileSize:         document.FileSize,
		Status:           string(document.Status),
		ProcessingStatus: string(document.ProcessingStatus),
		CreatedAt:        document.CreatedAt,
		UpdatedAt:        document.UpdatedAt,
	}
}

func buildObjectKey(categoryName, originalName string) string {
	safeCategory := sanitizeSegment(strings.ToLower(strings.TrimSpace(categoryName)))
	safeFileName := sanitizeFileName(originalName)
	return "documents/" + safeCategory + "/" + randomID() + "-" + safeFileName
}

func sanitizeSegment(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "..", "")
	return replacer.Replace(input)
}

func sanitizeFileName(filename string) string {
	filename = filepath.Base(strings.TrimSpace(filename))
	if filename == "" || filename == "." || filename == ".." {
		return "file"
	}

	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		"..", "_",
	)
	filename = replacer.Replace(filename)
	if filename == "" {
		return "file"
	}
	return filename
}

func extractFileName(objectKey string) string {
	base := filepath.Base(strings.TrimSpace(objectKey))
	if base == "" || base == "." || base == ".." {
		return ""
	}
	if idx := strings.Index(base, "-"); idx >= 0 && idx+1 < len(base) {
		return base[idx+1:]
	}
	return base
}

func detectContentType(fileHeader *multipart.FileHeader) string {
	contentType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
	if contentType == "" {
		return "application/octet-stream"
	}
	return contentType
}

func randomID() string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return time.Now().UTC().Format("20060102150405")
	}
	return hex.EncodeToString(bytes)
}

func (s *documentService) scheduleRAGIndexing(ctx context.Context, document *doccategory.Document) {
	if s.ragService == nil {
		return
	}
	clone := *document
	go func() {
		_ = s.ragService.IndexDocument(ctx, &clone)
	}()
}
