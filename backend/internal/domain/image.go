package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// MaxImageSizeBytes is the maximum allowed file size for images (10MB).
const MaxImageSizeBytes = 10 * 1024 * 1024

// ContentType represents the MIME type of an image.
type ContentType string

const (
	ContentTypeJPEG ContentType = "image/jpeg"
	ContentTypePNG  ContentType = "image/png"
)

// ImageType represents the purpose of an image in the production process.
type ImageType string

const (
	ImageTypeProcess  ImageType = "process"  // 工程記録
	ImageTypeMacro    ImageType = "macro"    // マクロ撮影
	ImageTypeAging    ImageType = "aging"    // 経年変化
	ImageTypeOverview ImageType = "overview" // 全体写真
)

// Image represents a photograph attached to a work or process step.
type Image struct {
	ID            uuid.UUID   `json:"id"`
	WorkID        uuid.UUID   `json:"work_id"`
	ProcessStepID *uuid.UUID  `json:"process_step_id,omitempty"`
	FilePath      string      `json:"file_path"`
	FileSizeBytes int64       `json:"file_size_bytes"`
	ContentType   ContentType `json:"content_type"`
	ImageType     ImageType   `json:"image_type"`
	Caption       string      `json:"caption,omitempty"`
	TakenAt       *time.Time  `json:"taken_at,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
}

// Validate checks the image's fields for correctness.
func (img *Image) Validate() error {
	if img.WorkID == uuid.Nil {
		return errors.New("work_id must not be empty")
	}
	if img.FilePath == "" {
		return errors.New("file_path must not be empty")
	}
	if img.FileSizeBytes <= 0 {
		return errors.New("file_size_bytes must be positive")
	}
	if img.FileSizeBytes > MaxImageSizeBytes {
		return errors.New("file_size_bytes must not exceed 10MB")
	}
	if !isValidContentType(img.ContentType) {
		return errors.New("content_type must be image/jpeg or image/png")
	}
	if !isValidImageType(img.ImageType) {
		return errors.New("invalid image_type")
	}
	if len(img.Caption) > 500 {
		return errors.New("caption must not exceed 500 characters")
	}
	return nil
}

func isValidContentType(ct ContentType) bool {
	switch ct {
	case ContentTypeJPEG, ContentTypePNG:
		return true
	}
	return false
}

func isValidImageType(it ImageType) bool {
	switch it {
	case ImageTypeProcess, ImageTypeMacro, ImageTypeAging, ImageTypeOverview:
		return true
	}
	return false
}
