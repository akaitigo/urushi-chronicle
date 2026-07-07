package storage_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/akaitigo/urushi-chronicle/internal/storage"
	"github.com/google/uuid"
)

func discardLogger() *log.Logger {
	return log.New(os.Stderr, "", 0)
}

// testPrivateKeyPEM generates a throwaway RSA key encoded as a PKCS#8 PEM,
// matching the format of a real service account private_key field.
func testPrivateKeyPEM(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("failed to marshal key: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

// --- MockUploader ---

func TestMockUploader_GenerateUploadURL_JPEG(t *testing.T) {
	u := storage.NewMockUploader("test-bucket")
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

func TestMockUploader_GenerateUploadURL_PNG(t *testing.T) {
	u := storage.NewMockUploader("")
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

func TestMockUploader_GenerateUploadURL_UnsupportedType(t *testing.T) {
	u := storage.NewMockUploader("test-bucket")
	_, err := u.GenerateUploadURL(uuid.New(), uuid.New(), "image/gif")
	if err == nil {
		t.Error("expected error for unsupported content type")
	}
}

func TestMockUploader_GenerateUploadURL_UniqueFilePaths(t *testing.T) {
	u := storage.NewMockUploader("test-bucket")
	workID := uuid.New()
	stepID := uuid.New()

	result1, _ := u.GenerateUploadURL(workID, stepID, "image/jpeg")
	result2, _ := u.GenerateUploadURL(workID, stepID, "image/jpeg")

	if result1.FilePath == result2.FilePath {
		t.Error("each upload should generate a unique file path")
	}
}

// --- GCSUploader (real V4 signing, verified offline with a test key) ---

func TestGCSUploader_GenerateUploadURL_SignsV4(t *testing.T) {
	u, err := storage.NewGCSUploader("real-bucket", "svc@example.iam.gserviceaccount.com", testPrivateKeyPEM(t))
	if err != nil {
		t.Fatalf("NewGCSUploader: %v", err)
	}

	workID := uuid.New()
	stepID := uuid.New()
	result, err := u.GenerateUploadURL(workID, stepID, "image/jpeg")
	if err != nil {
		t.Fatalf("GenerateUploadURL: %v", err)
	}

	if !strings.Contains(result.UploadURL, "X-Goog-Algorithm=GOOG4-RSA-SHA256") {
		t.Errorf("expected a V4 signed URL, got %s", result.UploadURL)
	}
	if !strings.Contains(result.UploadURL, "X-Goog-Signature=") {
		t.Errorf("expected signature query parameter, got %s", result.UploadURL)
	}
	if !strings.Contains(result.UploadURL, "real-bucket") {
		t.Errorf("expected bucket in URL, got %s", result.UploadURL)
	}
	if !strings.HasSuffix(result.FilePath, ".jpg") {
		t.Errorf("expected .jpg path, got %s", result.FilePath)
	}
	if result.ExpiresAt.IsZero() {
		t.Error("expected non-zero expiry")
	}
}

func TestGCSUploader_GenerateUploadURL_UnsupportedType(t *testing.T) {
	u, err := storage.NewGCSUploader("real-bucket", "svc@example.iam.gserviceaccount.com", testPrivateKeyPEM(t))
	if err != nil {
		t.Fatalf("NewGCSUploader: %v", err)
	}
	if _, err := u.GenerateUploadURL(uuid.New(), uuid.New(), "image/gif"); err == nil {
		t.Error("expected error for unsupported content type")
	}
}

func TestNewGCSUploader_RequiresConfig(t *testing.T) {
	tests := []struct {
		name   string
		bucket string
		email  string
		key    []byte
	}{
		{"missing bucket", "", "svc@example.com", []byte("key")},
		{"missing email", "bucket", "", []byte("key")},
		{"missing key", "bucket", "svc@example.com", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := storage.NewGCSUploader(tt.bucket, tt.email, tt.key); err == nil {
				t.Error("expected error for incomplete config")
			}
		})
	}
}

// --- NewUploaderFromEnv ---

func TestNewUploaderFromEnv_MockMode(t *testing.T) {
	t.Setenv("STORAGE_MODE", "mock")
	t.Setenv("GCS_BUCKET", "dev-bucket")

	u, err := storage.NewUploaderFromEnv(discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := u.GenerateUploadURL(uuid.New(), uuid.New(), "image/png")
	if err != nil {
		t.Fatalf("GenerateUploadURL: %v", err)
	}
	if !strings.Contains(result.UploadURL, "dev-bucket") {
		t.Errorf("expected dev-bucket in mock URL, got %s", result.UploadURL)
	}
	if !strings.Contains(result.UploadURL, "mock-development-only") {
		t.Errorf("expected mock placeholder signature, got %s", result.UploadURL)
	}
}

func TestNewUploaderFromEnv_ProductionRequiresCredentials(t *testing.T) {
	t.Setenv("STORAGE_MODE", "")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")

	if _, err := storage.NewUploaderFromEnv(discardLogger()); err == nil {
		t.Error("expected error when credentials are missing in production mode")
	}
}

func TestNewUploaderFromEnv_LoadsServiceAccount(t *testing.T) {
	dir := t.TempDir()
	keyPEM := testPrivateKeyPEM(t)
	saJSON := `{"client_email":"svc@example.iam.gserviceaccount.com","private_key":` +
		strconv.Quote(string(keyPEM)) + `}`
	path := filepath.Join(dir, "sa.json")
	if err := os.WriteFile(path, []byte(saJSON), 0o600); err != nil {
		t.Fatalf("write sa.json: %v", err)
	}

	t.Setenv("STORAGE_MODE", "")
	t.Setenv("GCS_BUCKET", "prod-bucket")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", path)

	u, err := storage.NewUploaderFromEnv(discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := u.GenerateUploadURL(uuid.New(), uuid.New(), "image/jpeg")
	if err != nil {
		t.Fatalf("GenerateUploadURL: %v", err)
	}
	if !strings.Contains(result.UploadURL, "X-Goog-Algorithm=GOOG4-RSA-SHA256") {
		t.Errorf("expected real V4 signed URL, got %s", result.UploadURL)
	}
	if !strings.Contains(result.UploadURL, "prod-bucket") {
		t.Errorf("expected prod-bucket in URL, got %s", result.UploadURL)
	}
}

func TestNewUploaderFromEnv_InvalidCredentialsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("{ not valid json"), 0o600); err != nil {
		t.Fatalf("write bad.json: %v", err)
	}

	t.Setenv("STORAGE_MODE", "")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", path)

	if _, err := storage.NewUploaderFromEnv(discardLogger()); err == nil {
		t.Error("expected error for invalid credentials JSON")
	}
}
