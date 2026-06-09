// Package storage provides persistent storage for recall commands using SQLite.
//
// The database is stored at ~/.local/share/recall/recall.db by default.
// Each command is uniquely identified by its pattern (shell command template).
package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // SQLite driver — this import registers the "sqlite" driver with database/sql
)

// SQLiteStore implements the Store interface using an embedded SQLite database.
// It manages a single "commands" table that stores shell command patterns
// with metadata like tags, descriptions, usage counts, and workspace filters.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a store at the default location (~/.local/share/recall/recall.db).
// Creates the directory and database file if they don't exist.
func NewSQLiteStore() (*SQLiteStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	dbPath := filepath.Join(home, ".local", "share", "recall", "recall.db")
	return NewSQLiteStoreAt(dbPath)
}

// NewSQLiteStoreAt creates a store at the given file path.
// Creates parent directories if needed, opens the database, verifies connectivity,
// and runs schema migrations (create tables + indexes).
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
	if err := store.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return store, nil
}

// createTables initializes the database schema and runs migrations.
// Uses IF NOT EXISTS so it's safe to call on every startup.
func (s *SQLiteStore) createTables() error {
	// Core commands table (tags column kept for backward compat during migration)
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS commands (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pattern TEXT NOT NULL UNIQUE,
		description TEXT,
		tags TEXT,
		workspace_filter TEXT,
		source TEXT DEFAULT 'local',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME,
		usage_count INTEGER DEFAULT 0
	);`)
	if err != nil {
		return err
	}

	// Tags table: normalized, unique tag names
	_, err = s.db.Exec(`
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);`)
	if err != nil {
		return err
	}

	// Join table: many-to-many relationship between commands and tags
	_, err = s.db.Exec(`
	CREATE TABLE IF NOT EXISTS command_tags (
		command_id INTEGER NOT NULL REFERENCES commands(id) ON DELETE CASCADE,
		tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
		PRIMARY KEY (command_id, tag_id)
	);`)
	if err != nil {
		return err
	}

	// Migration: UNIQUE index for databases created before the constraint existed
	s.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_commands_pattern ON commands(pattern)")

	// Migration: move comma-separated tags from commands.tags column into the tags/command_tags tables
	if err := s.migrateTagsToTable(); err != nil {
		return err
	}

	return nil
}

// migrateTagsToTable moves any comma-separated tags from the legacy commands.tags column
// into the normalized tags + command_tags tables. Idempotent — skips commands already migrated.
func (s *SQLiteStore) migrateTagsToTable() error {
	rows, err := s.db.Query("SELECT id, tags FROM commands WHERE tags IS NOT NULL AND tags != ''")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cmdID int
		var tagsStr string
		if err := rows.Scan(&cmdID, &tagsStr); err != nil {
			return err
		}
		for _, tag := range splitTags(tagsStr) {
			if tag == "" {
				continue
			}
			s.db.Exec("INSERT OR IGNORE INTO tags (name) VALUES (?)", tag)
			s.db.Exec("INSERT OR IGNORE INTO command_tags (command_id, tag_id) SELECT ?, id FROM tags WHERE name = ?", cmdID, tag)
		}
	}
	return nil
}

// splitTags splits a comma-separated tag string into trimmed, lowercase tag names.
func splitTags(tags string) []string {
	var result []string
	for _, t := range strings.Split(tags, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, strings.ToLower(t))
		}
	}
	return result
}

// commandColumns is the SELECT column list used by all query methods.
// Must match the field order expected by scanCommandFrom.
const commandColumns = "id, pattern, description, tags, workspace_filter, source, usage_count, last_used_at"

// Close closes the underlying database connection.
// Should be called when the store is no longer needed (typically via defer).
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// List returns all stored commands, ordered by usage count (most used first),
// then by ID descending (newest first) as a tiebreaker.
// Returns an empty slice (not nil) if no commands exist.
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

// GetByID returns the command with the given ID, or nil if not found.
// Returns (nil, nil) when the ID doesn't exist — not an error.
func (s *SQLiteStore) GetByID(id int) (*Command, error) {
	row := s.db.QueryRow(
		"SELECT "+commandColumns+" FROM commands WHERE id = ?", id)
	cmd, err := scanCommandFrom(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

// GetByPattern returns the command matching the exact pattern string, or nil if not found.
// Pattern is unique, so at most one result is returned.
// Returns (nil, nil) when no match — not an error.
func (s *SQLiteStore) GetByPattern(pattern string) (*Command, error) {
	row := s.db.QueryRow(
		"SELECT "+commandColumns+" FROM commands WHERE pattern = ?", pattern)
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
// Called each time a user selects and executes a recalled command.
func (s *SQLiteStore) RecordUsage(id int) error {
	_, err := s.db.Exec(
		"UPDATE commands SET usage_count = usage_count + 1, last_used_at = ? WHERE id = ?",
		time.Now(), id,
	)
	return err
}

// Delete removes the command with the given ID.
// No error if the ID doesn't exist (DELETE WHERE is a no-op).
func (s *SQLiteStore) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM commands WHERE id = ?", id)
	return err
}

// Update replaces all mutable fields of the command identified by cmd.ID.
// Updates: pattern, description, tags, workspace_filter, updated_at.
// Does not modify usage_count, last_used_at, or source.
// Writes tags to both the legacy column and the normalized tags/command_tags tables.
func (s *SQLiteStore) Update(cmd Command) error {
	_, err := s.db.Exec(
		"UPDATE commands SET pattern = ?, description = ?, tags = ?, workspace_filter = ?, updated_at = ? WHERE id = ?",
		cmd.Pattern, cmd.Description, cmd.Tags, cmd.WorkspaceFilter, time.Now(), cmd.ID,
	)
	if err != nil {
		return err
	}
	return s.syncTags(cmd.ID, cmd.Tags)
}

// Upsert inserts a new command or updates an existing one if the pattern already exists.
// Uses SQLite's ON CONFLICT(pattern) to perform an atomic insert-or-update in a single statement.
// On conflict, updates description, tags, workspace_filter, source, and updated_at.
// Preserves the original created_at, usage_count, and last_used_at of existing commands.
// Defaults source to "local" if not specified.
// Writes tags to both the legacy column and the normalized tags/command_tags tables.
func (s *SQLiteStore) Upsert(cmd Command) error {
	source := cmd.Source
	if source == "" {
		source = "local"
	}

	_, err := s.db.Exec(`
		INSERT INTO commands (pattern, description, tags, workspace_filter, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(pattern) DO UPDATE SET
			description = excluded.description,
			tags = excluded.tags,
			workspace_filter = excluded.workspace_filter,
			source = excluded.source,
			updated_at = CURRENT_TIMESTAMP
	`, cmd.Pattern, cmd.Description, cmd.Tags, cmd.WorkspaceFilter, source)
	if err != nil {
		return err
	}

	// Look up the command ID (needed for upsert case where we don't know the ID)
	row := s.db.QueryRow("SELECT id FROM commands WHERE pattern = ?", cmd.Pattern)
	var cmdID int
	if err := row.Scan(&cmdID); err != nil {
		return err
	}
	return s.syncTags(cmdID, cmd.Tags)
}

// syncTags replaces all tag associations for a command.
// Clears existing command_tags rows, then inserts fresh ones from the comma-separated string.
// Creates new tag rows as needed via INSERT OR IGNORE.
func (s *SQLiteStore) syncTags(commandID int, tagsStr string) error {
	s.db.Exec("DELETE FROM command_tags WHERE command_id = ?", commandID)

	for _, tag := range splitTags(tagsStr) {
		if tag == "" {
			continue
		}
		s.db.Exec("INSERT OR IGNORE INTO tags (name) VALUES (?)", tag)
		s.db.Exec(
			"INSERT OR IGNORE INTO command_tags (command_id, tag_id) SELECT ?, id FROM tags WHERE name = ?",
			commandID, tag,
		)
	}
	return nil
}

// scanner is the common interface between sql.Row and sql.Rows — both have a Scan method.
type scanner interface {
	Scan(dest ...any) error
}

// scanCommandFrom scans a row into a Command.
// Handles nullable columns (description, tags, workspace_filter, source, last_used_at)
// by using sql.NullString/sql.NullTime and converting to Go zero values when NULL.
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
