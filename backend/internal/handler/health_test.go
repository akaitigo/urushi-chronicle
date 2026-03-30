package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akaitigo/urushi-chronicle/internal/handler"
)

func TestHealthHandler(t *testing.T) {
	h := handler.HealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}
