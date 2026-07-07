package migrate_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/akaitigo/urushi-chronicle/internal/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestDiscover_SortsAndFiltersUpFiles(t *testing.T) {
	dir := t.TempDir()
	// Intentionally out of order, plus non-migration files that must be ignored.
	writeFile(t, dir, "002_add_index.up.sql", "SELECT 2;")
	writeFile(t, dir, "001_create_core_tables.up.sql", "SELECT 1;")
	writeFile(t, dir, "001_create_core_tables.down.sql", "DROP;")
	writeFile(t, dir, "README.md", "docs")

	migrations, err := migrate.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(migrations) != 2 {
		t.Fatalf("expected 2 up-migrations, got %d", len(migrations))
	}
	if migrations[0].Version != "001" || migrations[0].Name != "001_create_core_tables" {
		t.Errorf("first migration = %+v, want version 001", migrations[0])
	}
	if migrations[1].Version != "002" || migrations[1].Name != "002_add_index" {
		t.Errorf("second migration = %+v, want version 002", migrations[1])
	}
	if migrations[0].SQL != "SELECT 1;" {
		t.Errorf("unexpected SQL for first migration: %q", migrations[0].SQL)
	}
}

func TestDiscover_RejectsInvalidName(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "not-a-valid-name.up.sql", "SELECT 1;")

	if _, err := migrate.Discover(dir); err == nil {
		t.Error("expected error for invalid migration file name")
	}
}

func TestDiscover_MissingDir(t *testing.T) {
	if _, err := migrate.Discover(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Error("expected error for missing migrations dir")
	}
}

func TestDiscover_EmptyDir(t *testing.T) {
	migrations, err := migrate.Discover(t.TempDir())
	if err != nil {
		t.Fatalf("Discover on empty dir: %v", err)
	}
	if len(migrations) != 0 {
		t.Errorf("expected 0 migrations, got %d", len(migrations))
	}
}

// TestRun_Integration exercises the runner against a real database. It is skipped
// unless DATABASE_URL is set (e.g. in CI). It uses a high version number and
// cleans up after itself so it never collides with the application's migrations.
func TestRun_Integration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping migration integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Skipf("failed to create pool (DB not available): %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("failed to ping database (DB not available): %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS migrate_selftest")
		_, _ = pool.Exec(ctx, "DELETE FROM schema_migrations WHERE version = '900'")
	})
	// Start clean in case a previous failed run left artifacts.
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS migrate_selftest")
	_, _ = pool.Exec(ctx, "DELETE FROM schema_migrations WHERE version = '900'")

	dir := t.TempDir()
	writeFile(t, dir, "900_migrate_selftest.up.sql",
		"CREATE TABLE migrate_selftest (id INT PRIMARY KEY);\nINSERT INTO migrate_selftest (id) VALUES (1);")

	// First run applies the migration.
	applied, err := migrate.Run(ctx, pool, dir)
	if err != nil {
		t.Fatalf("Run (first): %v", err)
	}
	if len(applied) != 1 || applied[0] != "900_migrate_selftest" {
		t.Fatalf("expected [900_migrate_selftest] applied, got %v", applied)
	}

	// The table and its row must exist (proves the multi-statement SQL ran).
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM migrate_selftest").Scan(&count); err != nil {
		t.Fatalf("query migrate_selftest: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row in migrate_selftest, got %d", count)
	}

	// Second run is a no-op (idempotent).
	appliedAgain, err := migrate.Run(ctx, pool, dir)
	if err != nil {
		t.Fatalf("Run (second): %v", err)
	}
	if len(appliedAgain) != 0 {
		t.Errorf("expected no migrations on second run, got %v", appliedAgain)
	}
}
