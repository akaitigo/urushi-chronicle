// Package storage defines interfaces and implementations for file storage.
package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PresignedURL holds the upload URL and metadata for a file upload.
type PresignedURL struct {
	UploadURL   string    `json:"upload_url"`
	FilePath    string    `json:"file_path"`
	ContentType string    `json:"content_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// ImageUploader generates presigned URLs for image uploads.
type ImageUploader interface {
	GenerateUploadURL(workID, stepID uuid.UUID, contentType string) (*PresignedURL, error)
}

// GCSUploader generates presigned upload URLs for Google Cloud Storage.
// In the MVP implementation, it generates deterministic paths without actual GCS signing.
type GCSUploader struct {
	bucketName string
	baseURL    string
}

// NewGCSUploader creates a new GCSUploader. If bucketName is empty, it defaults to "urushi-chronicle-dev".
func NewGCSUploader(bucketName string) *GCSUploader {
	if bucketName == "" {
		bucketName = "urushi-chronicle-dev"
	}
	return &GCSUploader{
		bucketName: bucketName,
		baseURL:    fmt.Sprintf("https://storage.googleapis.com/%s", bucketName),
	}
}

// allowedContentTypes maps permitted MIME types to file extensions.
var allowedContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
}

// GenerateUploadURL creates a presigned URL for uploading an image.
// The file is stored at: works/{workID}/steps/{stepID}/{uuid}{ext}
func (g *GCSUploader) GenerateUploadURL(workID, stepID uuid.UUID, contentType string) (*PresignedURL, error) {
	ext, ok := allowedContentTypes[contentType]
	if !ok {
		return nil, errors.New("unsupported content_type: must be image/jpeg or image/png")
	}

	fileID := uuid.New()
	filePath := fmt.Sprintf("works/%s/steps/%s/%s%s", workID, stepID, fileID, ext)

	return &PresignedURL{
		UploadURL:   fmt.Sprintf("%s/%s?X-Goog-Signature=mock-presigned-token", g.baseURL, filePath),
		FilePath:    filePath,
		ContentType: contentType,
		ExpiresAt:   time.Now().Add(15 * time.Minute),
	}, nil
}
