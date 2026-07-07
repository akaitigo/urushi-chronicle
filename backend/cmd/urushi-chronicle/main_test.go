package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestPackageImports(t *testing.T) {
	// Verify that the main package compiles without error.
	// The actual main() starts an HTTP server and blocks,
	// so we only validate that the package builds successfully.
	t.Log("main package compiles OK")
}

func TestRunServer_ShutsDownOnContextCancel(t *testing.T) {
	srv := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: http.NewServeMux(),
	}
	ctx, cancel := context.WithCancel(context.Background())
	logger := log.New(io.Discard, "", 0)

	done := make(chan error, 1)
	go func() {
		done <- runServer(ctx, srv, 5*time.Second, logger)
	}()

	// Give the server a moment to start listening, then request shutdown.
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected clean graceful shutdown, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("runServer did not return after context cancellation")
	}
}

func TestRunServer_ReturnsServeError(t *testing.T) {
	// Port 99999 is out of the valid range, so ListenAndServe fails immediately.
	srv := &http.Server{
		Addr:    "127.0.0.1:99999",
		Handler: http.NewServeMux(),
	}
	logger := log.New(io.Discard, "", 0)

	err := runServer(context.Background(), srv, time.Second, logger)
	if err == nil {
		t.Fatal("expected an error when the server fails to listen")
	}
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
