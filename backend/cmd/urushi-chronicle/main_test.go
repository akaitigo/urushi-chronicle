package main

import (
	"testing"
)

func TestPackageImports(t *testing.T) {
	// Verify that the main package compiles without error.
	// The actual main() starts an HTTP server and blocks,
	// so we only validate that the package builds successfully.
	t.Log("main package compiles OK")
}
