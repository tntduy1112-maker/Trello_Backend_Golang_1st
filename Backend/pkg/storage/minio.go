package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageService interface {
	Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error)
	Delete(ctx context.Context, objectKey string) error
	GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
	GetPublicURL(objectKey string) string
}

type MinioStorage struct {
	client     *minio.Client
	bucket     string
	endpoint   string
	useSSL     bool
	publicHost string
}

type MinioConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Bucket     string
	UseSSL     bool
	PublicHost string
}

func NewMinioStorage(cfg MinioConfig) (*MinioStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	publicHost := cfg.PublicHost
	if publicHost == "" {
		protocol := "http"
		if cfg.UseSSL {
			protocol = "https"
		}
		publicHost = fmt.Sprintf("%s://%s/%s", protocol, cfg.Endpoint, cfg.Bucket)
	}

	return &MinioStorage{
		client:     client,
		bucket:     cfg.Bucket,
		endpoint:   cfg.Endpoint,
		useSSL:     cfg.UseSSL,
		publicHost: publicHost,
	}, nil
}

func (s *MinioStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object: %w", err)
	}

	return s.GetPublicURL(objectKey), nil
}

func (s *MinioStorage) Delete(ctx context.Context, objectKey string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
}

func (s *MinioStorage) GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, expiry, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (s *MinioStorage) GetPublicURL(objectKey string) string {
	return fmt.Sprintf("%s/%s", s.publicHost, objectKey)
}

// MockStorage for testing
type MockStorage struct {
	files map[string][]byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{files: make(map[string][]byte)}
}

func (s *MockStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	data, _ := io.ReadAll(reader)
	s.files[objectKey] = data
	return fmt.Sprintf("http://localhost:9000/test-bucket/%s", objectKey), nil
}

func (s *MockStorage) Delete(ctx context.Context, objectKey string) error {
	delete(s.files, objectKey)
	return nil
}

func (s *MockStorage) GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	return fmt.Sprintf("http://localhost:9000/test-bucket/%s?token=xxx", objectKey), nil
}

func (s *MockStorage) GetPublicURL(objectKey string) string {
	return fmt.Sprintf("http://localhost:9000/test-bucket/%s", objectKey)
}
