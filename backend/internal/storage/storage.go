// Package storage defines interfaces and implementations for image file storage.
//
// Production uses Google Cloud Storage V4 signed URLs (GCSUploader), signed
// locally with a service account key. Local development can opt into a
// non-functional mock implementation via STORAGE_MODE=mock; see
// NewUploaderFromEnv. The mock is never selected on the production code path.
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	gcs "cloud.google.com/go/storage"
	"github.com/google/uuid"
)

// uploadURLTTL is how long a generated presigned upload URL remains valid.
const uploadURLTTL = 15 * time.Minute

// defaultBucketName is used when GCS_BUCKET is not set.
const defaultBucketName = "urushi-chronicle-dev"

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

// ErrUnsupportedContentType is returned when a content type is not permitted.
var ErrUnsupportedContentType = errors.New("unsupported content_type: must be image/jpeg or image/png")

// allowedContentTypes maps permitted MIME types to file extensions.
var allowedContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
}

// objectPath builds the storage object path for an uploaded image:
//
//	works/{workID}/steps/{stepID}/{uuid}{ext}
//
// A fresh UUID is generated per call so repeated uploads never collide.
func objectPath(workID, stepID uuid.UUID, ext string) string {
	return fmt.Sprintf("works/%s/steps/%s/%s%s", workID, stepID, uuid.New(), ext)
}

// GCSUploader generates real V4 signed PUT URLs for Google Cloud Storage.
// URLs are signed locally with the service account private key, so no network
// round-trip is required when issuing an upload URL.
type GCSUploader struct {
	bucketName     string
	googleAccessID string
	privateKey     []byte
}

// NewGCSUploader creates a GCSUploader that signs upload URLs with the given
// service account credentials. All arguments are required.
func NewGCSUploader(bucketName, googleAccessID string, privateKey []byte) (*GCSUploader, error) {
	if bucketName == "" {
		return nil, errors.New("bucket name is required")
	}
	if googleAccessID == "" {
		return nil, errors.New("google access ID (service account email) is required")
	}
	if len(privateKey) == 0 {
		return nil, errors.New("signing private key is required")
	}
	return &GCSUploader{
		bucketName:     bucketName,
		googleAccessID: googleAccessID,
		privateKey:     privateKey,
	}, nil
}

// GenerateUploadURL creates a V4 signed PUT URL for uploading an image to GCS.
func (g *GCSUploader) GenerateUploadURL(workID, stepID uuid.UUID, contentType string) (*PresignedURL, error) {
	ext, ok := allowedContentTypes[contentType]
	if !ok {
		return nil, ErrUnsupportedContentType
	}

	filePath := objectPath(workID, stepID, ext)
	expires := time.Now().Add(uploadURLTTL)

	signedURL, err := gcs.SignedURL(g.bucketName, filePath, &gcs.SignedURLOptions{
		GoogleAccessID: g.googleAccessID,
		PrivateKey:     g.privateKey,
		Method:         http.MethodPut,
		Expires:        expires,
		ContentType:    contentType,
		Scheme:         gcs.SigningSchemeV4,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign GCS upload URL: %w", err)
	}

	return &PresignedURL{
		UploadURL:   signedURL,
		FilePath:    filePath,
		ContentType: contentType,
		ExpiresAt:   expires,
	}, nil
}

// NewUploaderFromEnv selects an ImageUploader based on environment configuration.
//
//   - STORAGE_MODE=mock              → MockUploader (local development only)
//   - otherwise (production default) → real GCS V4 signer, requiring
//     GOOGLE_APPLICATION_CREDENTIALS to point at a service account JSON key.
//
// The bucket name comes from GCS_BUCKET (default "urushi-chronicle-dev").
// The mock is only reachable when STORAGE_MODE=mock is set explicitly, so the
// production code path never returns mock signatures.
func NewUploaderFromEnv(logger *log.Logger) (ImageUploader, error) {
	bucketName := os.Getenv("GCS_BUCKET")
	if bucketName == "" {
		bucketName = defaultBucketName
	}

	if strings.EqualFold(os.Getenv("STORAGE_MODE"), "mock") {
		if logger != nil {
			logger.Println("storage: STORAGE_MODE=mock — generating MOCK upload URLs (not valid for real GCS uploads)")
		}
		return NewMockUploader(bucketName), nil
	}

	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsPath == "" {
		return nil, errors.New("GOOGLE_APPLICATION_CREDENTIALS must be set for GCS signing " +
			"(set STORAGE_MODE=mock for local development)")
	}

	email, privateKey, err := loadServiceAccountKey(credsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load GCS service account key: %w", err)
	}
	return NewGCSUploader(bucketName, email, privateKey)
}

// serviceAccountKey holds the fields required from a GCP service account JSON key.
type serviceAccountKey struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
}

// loadServiceAccountKey reads a service account JSON key file and returns the
// client email and PEM-encoded private key used for signing.
func loadServiceAccountKey(path string) (email string, privateKey []byte, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read credentials file: %w", err)
	}
	var key serviceAccountKey
	if err := json.Unmarshal(data, &key); err != nil {
		return "", nil, fmt.Errorf("failed to parse credentials JSON: %w", err)
	}
	if key.ClientEmail == "" || key.PrivateKey == "" {
		return "", nil, errors.New("credentials JSON missing client_email or private_key")
	}
	return key.ClientEmail, []byte(key.PrivateKey), nil
}
