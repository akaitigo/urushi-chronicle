package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/handler"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/akaitigo/urushi-chronicle/internal/storage"
	"github.com/google/uuid"
)

// setupTest creates a step handler with seeded test data and returns the handler + work ID.
func setupTest() (*handler.StepHandler, uuid.UUID) {
	workRepo := repository.NewMemoryWorkRepository()
	stepRepo := repository.NewMemoryStepRepository()
	uploader := storage.NewGCSUploader("test-bucket")

	workID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	workRepo.Seed(&domain.Work{
		ID:        workID,
		Title:     "テスト作品",
		Technique: domain.TechniqueMakie,
		Status:    domain.WorkStatusInProgress,
		StartedAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	h := handler.NewStepHandler(stepRepo, workRepo, uploader)
	return h, workID
}

func TestCreateStep_Success(t *testing.T) {
	h, workID := setupTest()

	body := `{"name":"下塗り一回目","step_order":1,"category":"shitanuri"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/works/"+workID.String()+"/steps", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["name"] != "下塗り一回目" {
		t.Errorf("expected name '下塗り一回目', got %v", resp["name"])
	}
}

func TestCreateStep_ValidationError(t *testing.T) {
	h, workID := setupTest()

	body := `{"name":"","step_order":1,"category":"shitanuri"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/works/"+workID.String()+"/steps", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateStep_WorkNotFound(t *testing.T) {
	h, _ := setupTest()
	fakeWorkID := uuid.New()

	body := `{"name":"テスト","step_order":1,"category":"shitanuri"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/works/"+fakeWorkID.String()+"/steps", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestListSteps_Empty(t *testing.T) {
	h, workID := setupTest()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/works/"+workID.String()+"/steps", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["total"].(float64) != 0 {
		t.Errorf("expected total 0, got %v", resp["total"])
	}
}

func createTestStep(t *testing.T, h *handler.StepHandler, workID uuid.UUID, name string, order int) string {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"name":       name,
		"step_order": order,
		"category":   "shitanuri",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/works/"+workID.String()+"/steps", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("failed to create step: %d %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	return resp["id"].(string)
}

func TestGetStep_Success(t *testing.T) {
	h, workID := setupTest()
	stepID := createTestStep(t, h, workID, "下塗り", 1)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/works/"+workID.String()+"/steps/"+stepID, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetStep_NotFound(t *testing.T) {
	h, workID := setupTest()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/works/"+workID.String()+"/steps/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateStep_Success(t *testing.T) {
	h, workID := setupTest()
	stepID := createTestStep(t, h, workID, "下塗り", 1)

	body := `{"name":"下塗り（修正版）"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/works/"+workID.String()+"/steps/"+stepID, bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["name"] != "下塗り（修正版）" {
		t.Errorf("expected updated name, got %v", resp["name"])
	}
}

func TestUpdateStep_NotFound(t *testing.T) {
	h, workID := setupTest()

	body := `{"name":"不明"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/works/"+workID.String()+"/steps/"+uuid.New().String(), bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDeleteStep_Success(t *testing.T) {
	h, workID := setupTest()
	stepID := createTestStep(t, h, workID, "下塗り", 1)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/works/"+workID.String()+"/steps/"+stepID, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	// Verify deleted
	req = httptest.NewRequest(http.MethodGet, "/api/v1/works/"+workID.String()+"/steps/"+stepID, nil)
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rr.Code)
	}
}

func TestDeleteStep_NotFound(t *testing.T) {
	h, workID := setupTest()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/works/"+workID.String()+"/steps/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUploadURL_Success(t *testing.T) {
	h, workID := setupTest()
	stepID := createTestStep(t, h, workID, "下塗り", 1)

	body := `{"content_type":"image/jpeg"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/works/"+workID.String()+"/steps/"+stepID+"/upload", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["upload_url"] == nil {
		t.Error("expected upload_url in response")
	}
	if resp["file_path"] == nil {
		t.Error("expected file_path in response")
	}
}

func TestUploadURL_StepNotFound(t *testing.T) {
	h, workID := setupTest()

	body := `{"content_type":"image/jpeg"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/works/"+workID.String()+"/steps/"+uuid.New().String()+"/upload", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUploadURL_UnsupportedContentType(t *testing.T) {
	h, workID := setupTest()
	stepID := createTestStep(t, h, workID, "下塗り", 1)

	body := `{"content_type":"image/gif"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/works/"+workID.String()+"/steps/"+stepID+"/upload", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestListSteps_AfterCreation(t *testing.T) {
	h, workID := setupTest()
	createTestStep(t, h, workID, "下塗り", 1)
	createTestStep(t, h, workID, "中塗り", 2)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/works/"+workID.String()+"/steps", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["total"].(float64) != 2 {
		t.Errorf("expected total 2, got %v", resp["total"])
	}
	items := resp["items"].([]any)
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestInvalidPath(t *testing.T) {
	h, _ := setupTest()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/invalid/path", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h, workID := setupTest()

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/works/"+workID.String()+"/steps", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
