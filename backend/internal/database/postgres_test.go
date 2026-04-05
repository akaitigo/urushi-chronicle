package database_test

import (
	"context"
	"testing"

	"github.com/akaitigo/urushi-chronicle/internal/database"
)

func TestNewPool_EmptyURL(t *testing.T) {
	ctx := context.Background()
	_, err := database.NewPool(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty DATABASE_URL, got nil")
	}
}

func TestNewPool_InvalidURL(t *testing.T) {
	ctx := context.Background()
	_, err := database.NewPool(ctx, "not-a-valid-url")
	if err == nil {
		t.Fatal("expected error for invalid DATABASE_URL, got nil")
	}
}
