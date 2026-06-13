// Package storage provides persistent storage for recall commands using SQLite.
//
// The database location is resolved by internal/paths (respects RECALL_DB_PATH
// and XDG_DATA_HOME). Each command is uniquely identified by its pattern.
package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/CognisiveLabs/recall-cli/internal/paths"
	_ "modernc.org/sqlite" // SQLite driver — registers the "sqlite" driver with database/sql
)

// SQLiteStore implements the Storage interface using an embedded SQLite database.
// It manages a single "commands" table with metadata like tags, descriptions,
// usage counts, and workspace filters.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) the database at the location resolved by
// internal/paths. Respects RECALL_DB_PATH and XDG_DATA_HOME env vars.
func NewSQLiteStore() (*SQLiteStore, error) {
	dbPath, err := paths.DBPath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve db path: %w", err)
	}
	return NewSQLiteStoreAt(dbPath)
}

// NewSQLiteStoreAt opens (or creates) the database at the given path.
// Creates parent directories as needed. Safe to call on every startup because
// the schema uses IF NOT EXISTS throughout.
// Use this constructor in tests to point at a temp file.
func NewSQLiteStoreAt(dbPath string) (*SQLiteStore, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialise schema: %w", err)
	}

	return store, nil
}

// initSchema creates the commands table and the unique pattern index.
// All statements are idempotent (IF NOT EXISTS), so this is safe to run on
// every startup against an existing database.
func (s *SQLiteStore) initSchema() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS commands (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			pattern          TEXT NOT NULL UNIQUE,
			description      TEXT,
			tags             TEXT,
			workspace_filter TEXT,
			source           TEXT DEFAULT 'local',
			created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_used_at     DATETIME,
			usage_count      INTEGER DEFAULT 0
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_commands_pattern ON commands(pattern);
	`)
	return err
}

// commandColumns is the SELECT column list shared by all query methods.
// Must match the field order expected by scanCommandFrom.
const commandColumns = "id, pattern, description, tags, workspace_filter, source, usage_count, last_used_at"

// Close closes the underlying database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// List returns all stored commands, ordered by usage count (most used first),
// then by ID descending (newest first) as a tiebreaker.
func (s *SQLiteStore) List() ([]Command, error) {
	rows, err := s.db.Query(
		"SELECT " + commandColumns + " FROM commands ORDER BY usage_count DESC, id DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []Command
	for rows.Next() {
		cmd, err := scanCommandFrom(rows)
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}

// GetByID returns the command with the given ID, or (nil, nil) if not found.
func (s *SQLiteStore) GetByID(id int) (*Command, error) {
	row := s.db.QueryRow("SELECT "+commandColumns+" FROM commands WHERE id = ?", id)
	cmd, err := scanCommandFrom(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

// GetByPattern returns the command with the exact pattern, or (nil, nil) if not found.
func (s *SQLiteStore) GetByPattern(pattern string) (*Command, error) {
	row := s.db.QueryRow("SELECT "+commandColumns+" FROM commands WHERE pattern = ?", pattern)
	cmd, err := scanCommandFrom(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

// RecordUsage increments the usage_count and updates last_used_at for the given command.
func (s *SQLiteStore) RecordUsage(id int) error {
	_, err := s.db.Exec(
		"UPDATE commands SET usage_count = usage_count + 1, last_used_at = ? WHERE id = ?",
		time.Now(), id,
	)
	return err
}

// Delete removes the command with the given ID. No-op if the ID does not exist.
func (s *SQLiteStore) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM commands WHERE id = ?", id)
	return err
}

// Update replaces the mutable fields (pattern, description, tags, workspace_filter)
// of the command identified by cmd.ID and bumps updated_at.
// Does not touch usage_count, last_used_at, or source.
func (s *SQLiteStore) Update(cmd Command) error {
	_, err := s.db.Exec(
		`UPDATE commands
		    SET pattern = ?, description = ?, tags = ?, workspace_filter = ?, updated_at = ?
		  WHERE id = ?`,
		cmd.Pattern, cmd.Description, cmd.Tags, cmd.WorkspaceFilter, time.Now(), cmd.ID,
	)
	return err
}

// Upsert inserts a new command or updates an existing one when the pattern already
// exists. On conflict, description, tags, workspace_filter, source, and updated_at
// are refreshed while created_at, usage_count, and last_used_at are preserved.
// Defaults source to "local" when not set.
func (s *SQLiteStore) Upsert(cmd Command) error {
	source := cmd.Source
	if source == "" {
		source = "local"
	}

	_, err := s.db.Exec(`
		INSERT INTO commands (pattern, description, tags, workspace_filter, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(pattern) DO UPDATE SET
			description      = excluded.description,
			tags             = excluded.tags,
			workspace_filter = excluded.workspace_filter,
			source           = excluded.source,
			updated_at       = CURRENT_TIMESTAMP
	`, cmd.Pattern, cmd.Description, cmd.Tags, cmd.WorkspaceFilter, source)
	return err
}


// scanner is the common interface for sql.Row and sql.Rows — both expose Scan.
type scanner interface {
	Scan(dest ...any) error
}

// scanCommandFrom reads a single row into a Command, handling nullable columns.
// Works with both sql.Row (single queries) and sql.Rows (list queries).
func scanCommandFrom(row scanner) (Command, error) {
	var cmd Command
	var desc, tags, workspace, source sql.NullString
	var lastUsed sql.NullTime

	err := row.Scan(&cmd.ID, &cmd.Pattern, &desc, &tags, &workspace, &source, &cmd.UsageCount, &lastUsed)
	if err != nil {
		return cmd, err
	}
	cmd.Description = desc.String
	cmd.Tags = tags.String
	cmd.WorkspaceFilter = workspace.String
	cmd.Source = source.String
	if lastUsed.Valid {
		cmd.LastUsedAt = lastUsed.Time
	}
	return cmd, nil
}
