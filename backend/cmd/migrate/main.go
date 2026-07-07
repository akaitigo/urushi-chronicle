// Command migrate applies pending SQL migrations from the migrations directory
// to the database identified by DATABASE_URL.
//
// Usage:
//
//	DATABASE_URL=postgres://... go run ./cmd/migrate
//
// Environment variables:
//
//	DATABASE_URL    (required) PostgreSQL connection string.
//	MIGRATIONS_DIR  (optional) path to *.up.sql files; defaults to "migrations".
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/database"
	"github.com/akaitigo/urushi-chronicle/internal/migrate"
)

func main() {
	logger := log.New(os.Stdout, "[migrate] ", log.LstdFlags)

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Fatal("DATABASE_URL must be set")
	}

	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "migrations"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		logger.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	applied, err := migrate.Run(ctx, pool, dir)
	if err != nil {
		logger.Fatalf("migration failed: %v", err)
	}

	if len(applied) == 0 {
		logger.Println("no pending migrations; schema is up to date")
		return
	}
	for _, name := range applied {
		logger.Printf("applied migration: %s", name)
	}
	logger.Printf("%d migration(s) applied", len(applied))
}
