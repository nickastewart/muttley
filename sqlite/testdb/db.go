// Package testdb opens a fresh SQLite file and applies the goose migrations
// under migrations/. Repository tests use it so each test gets its own database.
package testdb

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// Open returns a SQLite database backed by a new file in the test's temporary
// directory. The file is migrated with every goose Up section in migrations/.
func Open(t *testing.T) *sql.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	// One connection avoids SQLite's "database is locked" error when a test
	// uses the same file from more than one connection.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close sqlite database: %v", err)
		}
	})

	if err := applyMigrations(db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return db
}

func applyMigrations(db *sql.DB) error {
	dir, err := migrationsDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	if len(names) == 0 {
		return fmt.Errorf("no migration files in %s", dir)
	}

	for _, name := range names {
		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		statements, err := gooseUpStatements(string(body))
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		for _, statement := range statements {
			if _, err := db.Exec(statement); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	}
	return nil
}

// migrationsDir finds the repo migrations directory from this source file, then
// by walking up from the working directory. go test runs with the package
// directory as the working directory, so the source-file path is the usual hit.
func migrationsDir() (string, error) {
	var candidates []string
	if _, file, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Join(filepath.Dir(file), "..", "..", "migrations"))
	}
	if cwd, err := os.Getwd(); err == nil {
		for dir := cwd; ; dir = filepath.Dir(dir) {
			candidates = append(candidates, filepath.Join(dir, "migrations"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
		}
	}

	for _, dir := range candidates {
		info, err := os.Stat(filepath.Join(dir, "00001_create_user_table.sql"))
		if err == nil && !info.IsDir() {
			return dir, nil
		}
	}
	return "", fmt.Errorf("could not find migrations directory")
}

// gooseUpStatements returns the SQL inside each goose StatementBegin block in
// the Up section. Down sections are ignored. Every migration in this repo uses
// that form.
func gooseUpStatements(contents string) ([]string, error) {
	var statements []string
	inUp := false
	inStatement := false
	var b strings.Builder

	for _, line := range strings.Split(contents, "\n") {
		switch strings.TrimSpace(line) {
		case "-- +goose Up":
			inUp = true
			continue
		case "-- +goose Down":
			inUp = false
			inStatement = false
			continue
		case "-- +goose StatementBegin":
			if inUp {
				inStatement = true
				b.Reset()
			}
			continue
		case "-- +goose StatementEnd":
			if inUp && inStatement {
				statement := strings.TrimSpace(b.String())
				if statement != "" {
					statements = append(statements, statement)
				}
				inStatement = false
			}
			continue
		}
		if inUp && inStatement {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}

	if len(statements) == 0 {
		return nil, fmt.Errorf("migration has no goose Up statements")
	}
	return statements, nil
}
