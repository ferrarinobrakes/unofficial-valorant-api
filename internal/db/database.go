package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Database wraps the SQL database and SQLC queries
type Database struct {
	*Queries
	db *sql.DB
}

// New creates a new database connection
func New(dbPath string) (*Database, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set pragmas for performance
	pragmas := []string{
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = -64000", // 64MB cache
		"PRAGMA temp_store = MEMORY",
		"PRAGMA mmap_size = 268435456", // 256mb
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &Database{
		Queries: &Queries{db: db},
		db:      db,
	}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// runMigrations runs the SQL migrations
func runMigrations(db *sql.DB) error {
	schema, err := os.ReadFile("./sql/schema/001_initial.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}
