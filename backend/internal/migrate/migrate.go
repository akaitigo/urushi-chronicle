// Package migrate provides a minimal, forward-only SQL migration runner.
//
// Migrations are plain .up.sql files named "NNN_description.up.sql" (matching the
// golang-migrate convention already used in backend/migrations). Applied versions
// are tracked in a schema_migrations table so Run is idempotent. This intentionally
// avoids pulling in a heavyweight migration tool; see README for usage.
package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// upSuffix identifies forward ("up") migration files.
const upSuffix = ".up.sql"

// migrationNamePattern constrains migration base names to digits, lowercase
// letters, and underscores, so a recorded version can never contain unexpected input.
var migrationNamePattern = regexp.MustCompile(`^[0-9]+_[a-z0-9_]+$`)

// Migration is a single discovered up-migration.
type Migration struct {
	Version string // numeric prefix, e.g. "001"
	Name    string // base name without .up.sql, e.g. "001_create_core_tables"
	SQL     string
}

// Discover reads all *.up.sql files in dir, sorted lexicographically by file name.
func Discover(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations dir %q: %w", dir, err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), upSuffix) {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	migrations := make([]Migration, 0, len(names))
	for _, fileName := range names {
		base := strings.TrimSuffix(fileName, upSuffix)
		if !migrationNamePattern.MatchString(base) {
			return nil, fmt.Errorf("invalid migration file name %q: must match NNN_name%s", fileName, upSuffix)
		}
		sqlBytes, err := os.ReadFile(filepath.Join(dir, fileName))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %q: %w", fileName, err)
		}
		migrations = append(migrations, Migration{
			Version: base[:strings.IndexByte(base, '_')],
			Name:    base,
			SQL:     string(sqlBytes),
		})
	}
	return migrations, nil
}

// Run applies all pending migrations from dir against pool and returns the names
// of migrations that were applied (empty when the schema is already up to date).
func Run(ctx context.Context, pool *pgxpool.Pool, dir string) ([]string, error) {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return nil, fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}

	applied, err := appliedVersions(ctx, pool)
	if err != nil {
		return nil, err
	}

	migrations, err := Discover(dir)
	if err != nil {
		return nil, err
	}

	var ran []string
	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}
		if err := apply(ctx, pool, m); err != nil {
			return ran, fmt.Errorf("failed to apply migration %s: %w", m.Name, err)
		}
		ran = append(ran, m.Name)
	}
	return ran, nil
}

// appliedVersions returns the set of migration versions already recorded.
func appliedVersions(ctx context.Context, pool *pgxpool.Pool) (map[string]bool, error) {
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("failed to scan applied migration: %w", err)
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating applied migrations: %w", err)
	}
	return applied, nil
}

// apply executes a single migration's SQL, then records it.
//
// The SQL may contain multiple statements, so it is run via the simple query
// protocol (PgConn.Exec), which implicitly wraps a statement-only file in one
// transaction. The version/name are recorded with a parameterized INSERT.
func apply(ctx context.Context, pool *pgxpool.Pool, m Migration) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Conn().PgConn().Exec(ctx, m.SQL).ReadAll(); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	if _, err := conn.Exec(
		ctx,
		`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
		m.Version, m.Name,
	); err != nil {
		return fmt.Errorf("failed to record applied migration: %w", err)
	}
	return nil
}
