package db

import (
	"database/sql"
	"fmt"
)

// migration holds a versioned schema migration.
type migration struct {
	version int
	sql     string
}

// migrations is the ordered list of schema migrations.
// New migrations must be appended at the end with an incremented version.
var migrations = []migration{
	{
		version: 1,
		sql: `
CREATE TABLE cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    czech TEXT NOT NULL,
    english TEXT NOT NULL,
    deleted_at DATETIME NULL,
    suspended BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE card_tags (
    card_id INTEGER NOT NULL REFERENCES cards(id),
    tag TEXT NOT NULL,
    PRIMARY KEY (card_id, tag)
);

CREATE TABLE srs_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_id INTEGER NOT NULL REFERENCES cards(id),
    direction TEXT NOT NULL CHECK(direction IN ('cz_en', 'en_cz')),
    ease_factor REAL DEFAULT 2.5,
    interval_days REAL DEFAULT 0,
    repetitions INTEGER DEFAULT 0,
    next_review DATETIME DEFAULT CURRENT_TIMESTAMP,
    status TEXT DEFAULT 'new' CHECK(status IN ('new', 'learning', 'review')),
    learning_step INTEGER DEFAULT 0,
    UNIQUE(card_id, direction)
);

CREATE TABLE review_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    srs_state_id INTEGER NOT NULL REFERENCES srs_state(id),
    card_id INTEGER NOT NULL REFERENCES cards(id),
    direction TEXT NOT NULL,
    rating INTEGER NOT NULL CHECK(rating IN (1, 2, 3, 4)),
    reviewed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    interval_before REAL,
    interval_after REAL,
    ease_before REAL,
    ease_after REAL
);

CREATE INDEX idx_srs_next_review ON srs_state(next_review);
CREATE INDEX idx_srs_card_dir ON srs_state(card_id, direction);
CREATE INDEX idx_review_events_date ON review_events(reviewed_at);
CREATE INDEX idx_cards_deleted ON cards(deleted_at);
`,
	},
	{
		version: 2,
		sql: `
-- Remove en_cz direction: each card now has a single SRS state (cz_en only).
DELETE FROM review_events WHERE srs_state_id IN (SELECT id FROM srs_state WHERE direction = 'en_cz');
DELETE FROM srs_state WHERE direction = 'en_cz';
`,
	},
	{
		version: 3,
		sql: `
-- Re-create en_cz SRS states for all existing cards that lack one.
INSERT INTO srs_state (card_id, direction, ease_factor, interval_days, repetitions, next_review, status, learning_step)
SELECT c.id, 'en_cz', 2.5, 0, 0, CURRENT_TIMESTAMP, 'new', 0
FROM cards c
WHERE c.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM srs_state s WHERE s.card_id = c.id AND s.direction = 'en_cz');
`,
	},
	{
		version: 4,
		sql: `
-- Add full before-state columns to review_events for undo support.
ALTER TABLE review_events ADD COLUMN status_before TEXT;
ALTER TABLE review_events ADD COLUMN learning_step_before INTEGER;
ALTER TABLE review_events ADD COLUMN repetitions_before INTEGER;
ALTER TABLE review_events ADD COLUMN next_review_before DATETIME;
`,
	},
}

// RunMigrations applies all pending schema migrations to the database.
// It uses a schema_version table to track which migrations have been applied.
func RunMigrations(db *sql.DB) error {
	// Ensure the schema_version table exists.
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("creating schema_version table: %w", err)
	}

	// Determine the current schema version.
	var currentVersion int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("reading schema version: %w", err)
	}

	// Apply each pending migration inside a transaction.
	for _, m := range migrations {
		if m.version <= currentVersion {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("beginning transaction for migration %d: %w", m.version, err)
		}

		if _, err := tx.Exec(m.sql); err != nil {
			tx.Rollback()
			return fmt.Errorf("applying migration %d: %w", m.version, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (?)", m.version); err != nil {
			tx.Rollback()
			return fmt.Errorf("recording migration %d: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %d: %w", m.version, err)
		}
	}

	return nil
}
