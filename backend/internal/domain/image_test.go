package domain_test

import (
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
)

func validImage() domain.Image {
	return domain.Image{
		ID:            uuid.New(),
		WorkID:        uuid.New(),
		FilePath:      "works/abc123/step1.jpg",
		FileSizeBytes: 1024 * 1024, // 1MB
		ContentType:   domain.ContentTypeJPEG,
		ImageType:     domain.ImageTypeProcess,
		CreatedAt:     time.Now(),
	}
}

func TestImageValidate_Valid(t *testing.T) {
	img := validImage()
	if err := img.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestImageValidate_NilWorkID(t *testing.T) {
	img := validImage()
	img.WorkID = uuid.Nil
	if err := img.Validate(); err == nil {
		t.Error("expected error for nil work_id")
	}
}

func TestImageValidate_EmptyFilePath(t *testing.T) {
	img := validImage()
	img.FilePath = ""
	if err := img.Validate(); err == nil {
		t.Error("expected error for empty file_path")
	}
}

func TestImageValidate_ZeroFileSize(t *testing.T) {
	img := validImage()
	img.FileSizeBytes = 0
	if err := img.Validate(); err == nil {
		t.Error("expected error for zero file_size_bytes")
	}
}

func TestImageValidate_FileSizeTooLarge(t *testing.T) {
	img := validImage()
	img.FileSizeBytes = domain.MaxImageSizeBytes + 1
	if err := img.Validate(); err == nil {
		t.Error("expected error for file size exceeding 10MB")
	}
}

func TestImageValidate_ExactMaxFileSize(t *testing.T) {
	img := validImage()
	img.FileSizeBytes = domain.MaxImageSizeBytes
	if err := img.Validate(); err != nil {
		t.Errorf("exact max size should be valid, got %v", err)
	}
}

func TestImageValidate_InvalidContentType(t *testing.T) {
	img := validImage()
	img.ContentType = "image/gif"
	if err := img.Validate(); err == nil {
		t.Error("expected error for invalid content_type")
	}
}

func TestImageValidate_InvalidImageType(t *testing.T) {
	img := validImage()
	img.ImageType = "invalid"
	if err := img.Validate(); err == nil {
		t.Error("expected error for invalid image_type")
	}
}

func TestImageValidate_CaptionTooLong(t *testing.T) {
	img := validImage()
	img.Caption = string(make([]byte, 501))
	if err := img.Validate(); err == nil {
		t.Error("expected error for caption exceeding 500 characters")
	}
}

func TestImageValidate_AllContentTypes(t *testing.T) {
	types := []domain.ContentType{domain.ContentTypeJPEG, domain.ContentTypePNG}
	for _, ct := range types {
		img := validImage()
		img.ContentType = ct
		if err := img.Validate(); err != nil {
			t.Errorf("content type %s should be valid, got %v", ct, err)
		}
	}
}

func TestImageValidate_AllImageTypes(t *testing.T) {
	types := []domain.ImageType{
		domain.ImageTypeProcess,
		domain.ImageTypeMacro,
		domain.ImageTypeAging,
		domain.ImageTypeOverview,
	}
	for _, it := range types {
		img := validImage()
		img.ImageType = it
		if err := img.Validate(); err != nil {
			t.Errorf("image type %s should be valid, got %v", it, err)
		}
	}
}
