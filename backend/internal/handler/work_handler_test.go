package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestListWorks_Empty(t *testing.T) {
	h, _ := setupWorkTest()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/works", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
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

	var resp map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["total"].(float64) != 2 {
		t.Errorf("expected total 2, got %v", resp["total"])
	}
	items := resp["items"].([]interface{})
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

	var resp map[string]interface{}
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

func TestWorkHandler_MethodNotAllowed(t *testing.T) {
	h, _ := setupWorkTest()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/works", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
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

	var resp map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["total"].(float64) != 1 {
		t.Errorf("expected total 1, got %v", resp["total"])
	}
}
