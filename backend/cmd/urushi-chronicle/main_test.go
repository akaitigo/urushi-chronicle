package main

import (
	"testing"
)

// TestMain_Compiles verifies the main package compiles correctly.
// The actual main() starts an HTTP server, so we test the handlers separately.
func TestMain_Compiles(t *testing.T) {
	t.Run("main package compiles", func(t *testing.T) {
		// This test verifies compilation. Handler logic is tested in handler_test.go.
	})
}
