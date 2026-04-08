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

func TestValidateStoreType(t *testing.T) {
	tests := []struct {
		name      string
		storeType string
		wantErr   bool
	}{
		{name: "empty string is valid (defaults to memory)", storeType: "", wantErr: false},
		{name: "memory is valid", storeType: "memory", wantErr: false},
		{name: "postgres is valid", storeType: "postgres", wantErr: false},
		{name: "typo postgress is invalid", storeType: "postgress", wantErr: true},
		{name: "sqlite is invalid", storeType: "sqlite", wantErr: true},
		{name: "POSTGRES uppercase is invalid", storeType: "POSTGRES", wantErr: true},
		{name: "random string is invalid", storeType: "foobar", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStoreType(tt.storeType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStoreType(%q) error = %v, wantErr %v", tt.storeType, err, tt.wantErr)
			}
		})
	}
}
