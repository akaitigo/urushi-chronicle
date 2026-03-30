package storage_test

import (
	"strings"
	"testing"

	"github.com/akaitigo/urushi-chronicle/internal/storage"
	"github.com/google/uuid"
)

func TestGCSUploader_GenerateUploadURL_JPEG(t *testing.T) {
	u := storage.NewGCSUploader("test-bucket")
	workID := uuid.New()
	stepID := uuid.New()

	result, err := u.GenerateUploadURL(workID, stepID, "image/jpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.UploadURL, "test-bucket") {
		t.Errorf("upload URL should contain bucket name, got %s", result.UploadURL)
	}
	if !strings.Contains(result.FilePath, workID.String()) {
		t.Errorf("file path should contain workID, got %s", result.FilePath)
	}
	if !strings.Contains(result.FilePath, stepID.String()) {
		t.Errorf("file path should contain stepID, got %s", result.FilePath)
	}
	if !strings.HasSuffix(result.FilePath, ".jpg") {
		t.Errorf("JPEG file path should end with .jpg, got %s", result.FilePath)
	}
	if result.ContentType != "image/jpeg" {
		t.Errorf("expected content_type image/jpeg, got %s", result.ContentType)
	}
}

func TestGCSUploader_GenerateUploadURL_PNG(t *testing.T) {
	u := storage.NewGCSUploader("")
	result, err := u.GenerateUploadURL(uuid.New(), uuid.New(), "image/png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(result.FilePath, ".png") {
		t.Errorf("PNG file path should end with .png, got %s", result.FilePath)
	}
	if !strings.Contains(result.UploadURL, "urushi-chronicle-dev") {
		t.Errorf("default bucket should be urushi-chronicle-dev, got %s", result.UploadURL)
	}
}

func TestGCSUploader_GenerateUploadURL_UnsupportedType(t *testing.T) {
	u := storage.NewGCSUploader("test-bucket")
	_, err := u.GenerateUploadURL(uuid.New(), uuid.New(), "image/gif")
	if err == nil {
		t.Error("expected error for unsupported content type")
	}
}

func TestGCSUploader_GenerateUploadURL_UniqueFilePaths(t *testing.T) {
	u := storage.NewGCSUploader("test-bucket")
	workID := uuid.New()
	stepID := uuid.New()

	result1, _ := u.GenerateUploadURL(workID, stepID, "image/jpeg")
	result2, _ := u.GenerateUploadURL(workID, stepID, "image/jpeg")

	if result1.FilePath == result2.FilePath {
		t.Error("each upload should generate a unique file path")
	}
}
