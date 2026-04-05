package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/handler"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/google/uuid"
)

func setupWorkTest() (*handler.WorkHandler, *repository.MemoryWorkRepository) {
	workRepo := repository.NewMemoryWorkRepository()
	h := handler.NewWorkHandler(workRepo)
	return h, workRepo
}

func seedWork(repo *repository.MemoryWorkRepository, title string) uuid.UUID {
	id := uuid.New()
	now := time.Now().UTC()
	repo.Seed(&domain.Work{
		ID:        id,
		Title:     title,
		Technique: domain.TechniqueMakie,
		Status:    domain.WorkStatusInProgress,
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	})
	return id
}

// --- GET (existing tests) ---

func TestListWorks_Empty(t *testing.T) {
	h, _ := setupWorkTest()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/works", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["total"].(float64) != 0 {
		t.Errorf("expected total 0, got %v", resp["total"])
	}
}

func TestListWorks_WithData(t *testing.T) {
	h, repo := setupWorkTest()
	seedWork(repo, "蒔絵香合")
	seedWork(repo, "螺鈿硯箱")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/works", nil)
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

func TestGetWork_Success(t *testing.T) {
	h, repo := setupWorkTest()
	id := seedWork(repo, "蒔絵香合")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/works/"+id.String(), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["title"] != "蒔絵香合" {
		t.Errorf("expected title '蒔絵香合', got %v", resp["title"])
	}
}

func TestGetWork_NotFound(t *testing.T) {
	h, _ := setupWorkTest()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/works/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGetWork_InvalidID(t *testing.T) {
	h, _ := setupWorkTest()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/works/not-a-uuid", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestListWorks_TrailingSlash(t *testing.T) {
	h, repo := setupWorkTest()
	seedWork(repo, "テスト作品")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/works/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["total"].(float64) != 1 {
		t.Errorf("expected total 1, got %v", resp["total"])
	}
}

// --- POST (create) ---

func TestCreateWork_Success(t *testing.T) {
	h, _ := setupWorkTest()

	body := `{"title":"新規蒔絵作品","description":"テスト","technique":"makie","material":"欅"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/works", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["title"] != "新規蒔絵作品" {
		t.Errorf("expected title '新規蒔絵作品', got %v", resp["title"])
	}
	if resp["status"] != "in_progress" {
		t.Errorf("expected default status in_progress, got %v", resp["status"])
	}
	if resp["id"] == "" {
		t.Error("expected non-empty ID")
	}
}

func TestCreateWork_ValidationError_EmptyTitle(t *testing.T) {
	h, _ := setupWorkTest()

	body := `{"title":"","technique":"makie"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/works", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateWork_ValidationError_InvalidTechnique(t *testing.T) {
	h, _ := setupWorkTest()

	body := `{"title":"テスト","technique":"invalid"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/works", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateWork_InvalidJSON(t *testing.T) {
	h, _ := setupWorkTest()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/works", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// --- PUT (update) ---

func TestUpdateWork_Success(t *testing.T) {
	h, repo := setupWorkTest()
	id := seedWork(repo, "旧タイトル")

	body := `{"title":"新タイトル","technique":"raden"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/works/"+id.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["title"] != "新タイトル" {
		t.Errorf("expected '新タイトル', got %v", resp["title"])
	}
	if resp["technique"] != "raden" {
		t.Errorf("expected technique 'raden', got %v", resp["technique"])
	}
}

func TestUpdateWork_NotFound(t *testing.T) {
	h, _ := setupWorkTest()

	body := `{"title":"テスト"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/works/"+uuid.New().String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateWork_PartialUpdate(t *testing.T) {
	h, repo := setupWorkTest()
	id := seedWork(repo, "元のタイトル")

	// Only update status, title should remain
	body := `{"status":"completed"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/works/"+id.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["title"] != "元のタイトル" {
		t.Errorf("title should remain unchanged, got %v", resp["title"])
	}
	if resp["status"] != "completed" {
		t.Errorf("status should be 'completed', got %v", resp["status"])
	}
}

func TestUpdateWork_InvalidJSON(t *testing.T) {
	h, repo := setupWorkTest()
	id := seedWork(repo, "テスト")

	req := httptest.NewRequest(http.MethodPut, "/api/v1/works/"+id.String(), strings.NewReader("bad"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// --- DELETE ---

func TestDeleteWork_Success(t *testing.T) {
	h, repo := setupWorkTest()
	id := seedWork(repo, "削除対象")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/works/"+id.String(), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify deleted
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/works/"+id.String(), nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", getRR.Code)
	}
}

func TestDeleteWork_NotFound(t *testing.T) {
	h, _ := setupWorkTest()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/works/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// --- Method Not Allowed ---

func TestWorkHandler_MethodNotAllowed_Collection(t *testing.T) {
	h, _ := setupWorkTest()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/works", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestWorkHandler_MethodNotAllowed_Resource(t *testing.T) {
	h, _ := setupWorkTest()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/works/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
