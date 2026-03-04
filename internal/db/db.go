package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DefaultDBPath is the default database file location for local development.
const DefaultDBPath = "./data/flash-cards.db"

// DockerDBPath is the database file location when running inside Docker.
const DockerDBPath = "/data/flash-cards.db"

// Open opens a SQLite database at the given path with recommended settings:
// WAL journal mode, 5-second busy timeout, and foreign keys enabled.
// If dbPath is empty, DefaultDBPath is used.
// The parent directory is created if it does not exist.
// Schema migrations are applied automatically after opening.
func Open(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = DefaultDBPath
	}

	// Ensure the parent directory exists.
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating database directory %s: %w", dir, err)
	}

	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on", dbPath)

	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Verify the connection is usable.
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	// Apply schema migrations.
	if err := RunMigrations(sqlDB); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return sqlDB, nil
}
