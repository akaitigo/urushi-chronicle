package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akaitigo/urushi-chronicle/internal/handler"
)

func TestTemplateHandler_ListAll(t *testing.T) {
	h := handler.NewTemplateHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp) != 3 {
		t.Errorf("expected 3 templates, got %d", len(resp))
	}
}

func TestTemplateHandler_GetByName(t *testing.T) {
	h := handler.NewTemplateHandler()

	tests := []struct {
		path         string
		expectedName string
	}{
		{"/api/v1/templates/standard", "standard"},
		{"/api/v1/templates/makie", "makie"},
		{"/api/v1/templates/raden", "raden"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rr.Code)
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if resp["name"] != tt.expectedName {
				t.Errorf("expected name %q, got %v", tt.expectedName, resp["name"])
			}
		})
	}
}

func TestTemplateHandler_MethodNotAllowed(t *testing.T) {
	h := handler.NewTemplateHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
