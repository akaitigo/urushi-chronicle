// Package middleware provides HTTP middleware for the urushi-chronicle API.
package middleware

import (
	"net/http"
	"strings"
)

// CORS wraps an http.Handler and adds Cross-Origin Resource Sharing headers.
// allowedOrigins is a comma-separated list of origins (e.g. "http://localhost:3000,http://localhost:5173").
// If empty, no CORS headers are added (production behind same-origin reverse proxy).
func CORS(next http.Handler, allowedOrigins string) http.Handler {
	if allowedOrigins == "" {
		return next
	}

	origins := parseOrigins(allowedOrigins)
	originSet := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		originSet[o] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if _, ok := originSet[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400")
				w.Header().Set("Vary", "Origin")
			}
		}

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// parseOrigins splits a comma-separated origin string and trims whitespace.
func parseOrigins(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
