package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akaitigo/urushi-chronicle/pkg/middleware"
)

func dummyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestCORS_NoOriginConfigured(t *testing.T) {
	handler := middleware.CORS(dummyHandler(), "")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no CORS header when allowedOrigins is empty")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestCORS_AllowedOrigin(t *testing.T) {
	handler := middleware.CORS(dummyHandler(), "http://localhost:3000,http://localhost:5173")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("expected origin http://localhost:3000, got %q", got)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	handler := middleware.CORS(dummyHandler(), "http://localhost:3000")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS header for disallowed origin, got %q", got)
	}
}

func TestCORS_PreflightReturns204(t *testing.T) {
	handler := middleware.CORS(dummyHandler(), "http://localhost:3000")

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/works", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Allow-Methods header on preflight")
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	handler := middleware.CORS(dummyHandler(), "http://localhost:3000")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS header when no Origin sent, got %q", got)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestCORS_MultipleOriginsSecondAllowed(t *testing.T) {
	handler := middleware.CORS(dummyHandler(), "http://localhost:3000, http://localhost:5173")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("expected http://localhost:5173, got %q", got)
	}
}

func TestCORS_VaryHeaderSet(t *testing.T) {
	handler := middleware.CORS(dummyHandler(), "http://localhost:3000")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Vary"); got != "Origin" {
		t.Errorf("expected Vary: Origin, got %q", got)
	}
}
