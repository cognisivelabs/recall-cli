package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore() (*SQLiteStore, error) {
	// Determine database path (XDG compliant or default to home/recall)
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	dbDir := filepath.Join(home, ".local", "share", "recall")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db dir: %w", err)
	}

	dbPath := filepath.Join(dbDir, "recall.db")
	db, err := sql.Open("sqlite3", dbPath)
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

func (s *SQLiteStore) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS commands (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pattern TEXT NOT NULL,
		description TEXT,
		tags TEXT,
		workspace_filter TEXT,
		source TEXT DEFAULT 'local',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME,
		usage_count INTEGER DEFAULT 0
	);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) List() ([]Command, error) {
	rows, err := s.db.Query("SELECT id, pattern, description, tags, source FROM commands ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []Command
	for rows.Next() {
		var c Command
		// Handle NULL description/tags
		var desc, tags, source sql.NullString

		if err := rows.Scan(&c.ID, &c.Pattern, &desc, &tags, &source); err != nil {
			return nil, err
		}
		c.Description = desc.String
		c.Tags = tags.String
		c.Source = source.String

		cmds = append(cmds, c)
	}
	return cmds, nil
}

func (s *SQLiteStore) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM commands WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) Update(c Command) error {
	_, err := s.db.Exec("UPDATE commands SET pattern = ?, description = ?, tags = ?, updated_at = ? WHERE id = ?", c.Pattern, c.Description, c.Tags, time.Now(), c.ID)
	return err
}

func (s *SQLiteStore) Upsert(c Command) error {
	// Simple upsert logic: try to insert, if conflict (which we don't strictly enforce with constraints yet, but logic-wise)
	// For now, based on previous logic, this was separate handled in saveCommand.
	// But to satisfy interface, we can implement a basic Insert here.
	// NOTE: The previous logic in save command did a check first.
	// We will assume this is an Insert. The consumer decides Update vs Insert vs behavior.
	// Actually, let's just make this an INSERT. The TUI logic handles the "Check if exists" -> ID logic.
	_, err := s.db.Exec("INSERT INTO commands (pattern, description, tags, source, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		c.Pattern, c.Description, c.Tags, "local", time.Now(), time.Now())
	return err
}

func (s *SQLiteStore) GetByPattern(pattern string) (*Command, error) {
	var c Command
	var desc, tags, source sql.NullString

	err := s.db.QueryRow("SELECT id, pattern, description, tags, source FROM commands WHERE pattern = ?", pattern).Scan(&c.ID, &c.Pattern, &desc, &tags, &source)
	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, err
	}
	c.Description = desc.String
	c.Tags = tags.String
	c.Source = source.String

	return &c, nil
}
