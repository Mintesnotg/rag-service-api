package storage

import (
	"context"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStorage interface {
	Upload(ctx context.Context, objectKey string, file io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, objectKey string) error
	PresignedGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
}

type minioStorage struct {
	client *minio.Client
	bucket string
}

func NewMinIOStorage() (ObjectStorage, error) {
	endpoint, secure := resolveEndpoint()

	accessKey := strings.TrimSpace(os.Getenv("MINIO_ACCESS_KEY"))
	if accessKey == "" {
		accessKey = "minioadmin"
	}

	secretKey := strings.TrimSpace(os.Getenv("MINIO_SECRET_KEY"))
	if secretKey == "" {
		secretKey = "minioadmin"
	}

	bucket := strings.TrimSpace(os.Getenv("MINIO_BUCKET"))
	if bucket == "" {
		bucket = "docfiles"
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, err
	}

	s := &minioStorage{
		client: client,
		bucket: bucket,
	}

	if err := s.ensureBucket(context.Background()); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *minioStorage) Upload(ctx context.Context, objectKey string, file io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *minioStorage) Delete(ctx context.Context, objectKey string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
}

func (s *minioStorage) PresignedGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, expiry, nil)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

func (s *minioStorage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
}

func resolveEndpoint() (string, bool) {
	raw := strings.TrimSpace(os.Getenv("MINIO_ENDPOINT"))
	if raw == "" {
		raw = "http://127.0.0.1:9000"
	}

	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		if parsed, err := url.Parse(raw); err == nil && parsed.Host != "" {
			return parsed.Host, parsed.Scheme == "https"
		}
	}

	secure := strings.EqualFold(strings.TrimSpace(os.Getenv("MINIO_USE_SSL")), "true")
	return raw, secure
}
