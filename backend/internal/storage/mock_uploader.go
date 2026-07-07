package storage

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MockUploader generates deterministic, non-functional upload URLs for local
// development and testing. The returned URLs carry a placeholder signature and
// are NOT valid for real GCS uploads. Selection is gated behind STORAGE_MODE=mock
// in NewUploaderFromEnv so this implementation never reaches the production path.
type MockUploader struct {
	bucketName string
	baseURL    string
}

// NewMockUploader creates a MockUploader for the given bucket.
// If bucketName is empty, it defaults to defaultBucketName.
func NewMockUploader(bucketName string) *MockUploader {
	if bucketName == "" {
		bucketName = defaultBucketName
	}
	return &MockUploader{
		bucketName: bucketName,
		baseURL:    fmt.Sprintf("https://storage.googleapis.com/%s", bucketName),
	}
}

// GenerateUploadURL returns a deterministic mock presigned URL. The signature
// query parameter is a clearly-labeled placeholder, not a real credential.
func (m *MockUploader) GenerateUploadURL(workID, stepID uuid.UUID, contentType string) (*PresignedURL, error) {
	ext, ok := allowedContentTypes[contentType]
	if !ok {
		return nil, ErrUnsupportedContentType
	}

	filePath := objectPath(workID, stepID, ext)
	return &PresignedURL{
		UploadURL:   fmt.Sprintf("%s/%s?x-goog-signature=mock-development-only", m.baseURL, filePath),
		FilePath:    filePath,
		ContentType: contentType,
		ExpiresAt:   time.Now().Add(uploadURLTTL),
	}, nil
}
