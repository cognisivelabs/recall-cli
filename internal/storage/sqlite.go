package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a store at the default location (~/.local/share/recall/recall.db).
func NewSQLiteStore() (*SQLiteStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	dbPath := filepath.Join(home, ".local", "share", "recall", "recall.db")
	return NewSQLiteStoreAt(dbPath)
}

// NewSQLiteStoreAt creates a store at the given path. Useful for testing.
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
	rows, err := s.db.Query(`
		SELECT id, pattern, description, tags, workspace_filter, source, usage_count, last_used_at
		FROM commands
		ORDER BY usage_count DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []Command
	for rows.Next() {
		c, err := scanCommand(rows)
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, c)
	}
	return cmds, nil
}

func (s *SQLiteStore) GetByID(id int) (*Command, error) {
	row := s.db.QueryRow(`
		SELECT id, pattern, description, tags, workspace_filter, source, usage_count, last_used_at
		FROM commands WHERE id = ?`, id)
	c, err := scanCommandRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *SQLiteStore) GetByPattern(pattern string) (*Command, error) {
	row := s.db.QueryRow(`
		SELECT id, pattern, description, tags, workspace_filter, source, usage_count, last_used_at
		FROM commands WHERE pattern = ?`, pattern)
	c, err := scanCommandRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *SQLiteStore) RecordUsage(id int) error {
	_, err := s.db.Exec(
		"UPDATE commands SET usage_count = usage_count + 1, last_used_at = ? WHERE id = ?",
		time.Now(), id,
	)
	return err
}

func (s *SQLiteStore) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM commands WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) Update(c Command) error {
	_, err := s.db.Exec(
		"UPDATE commands SET pattern = ?, description = ?, tags = ?, workspace_filter = ?, updated_at = ? WHERE id = ?",
		c.Pattern, c.Description, c.Tags, c.WorkspaceFilter, time.Now(), c.ID,
	)
	return err
}

func (s *SQLiteStore) Upsert(c Command) error {
	source := c.Source
	if source == "" {
		source = "local"
	}

	existing, err := s.GetByPattern(c.Pattern)
	if err != nil {
		return err
	}

	if existing != nil {
		_, err = s.db.Exec(
			"UPDATE commands SET description = ?, tags = ?, workspace_filter = ?, source = ?, updated_at = ? WHERE id = ?",
			c.Description, c.Tags, c.WorkspaceFilter, source, time.Now(), existing.ID,
		)
		return err
	}

	_, err = s.db.Exec(
		"INSERT INTO commands (pattern, description, tags, workspace_filter, source, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		c.Pattern, c.Description, c.Tags, c.WorkspaceFilter, source, time.Now(), time.Now(),
	)
	return err
}

// scanCommand scans a row from sql.Rows into a Command.
func scanCommand(rows *sql.Rows) (Command, error) {
	var c Command
	var desc, tags, workspace, source sql.NullString
	var lastUsed sql.NullTime

	err := rows.Scan(&c.ID, &c.Pattern, &desc, &tags, &workspace, &source, &c.UsageCount, &lastUsed)
	if err != nil {
		return c, err
	}
	c.Description = desc.String
	c.Tags = tags.String
	c.WorkspaceFilter = workspace.String
	c.Source = source.String
	if lastUsed.Valid {
		c.LastUsedAt = lastUsed.Time
	}
	return c, nil
}

// scanCommandRow scans a single sql.Row into a Command.
func scanCommandRow(row *sql.Row) (Command, error) {
	var c Command
	var desc, tags, workspace, source sql.NullString
	var lastUsed sql.NullTime

	err := row.Scan(&c.ID, &c.Pattern, &desc, &tags, &workspace, &source, &c.UsageCount, &lastUsed)
	if err != nil {
		return c, err
	}
	c.Description = desc.String
	c.Tags = tags.String
	c.WorkspaceFilter = workspace.String
	c.Source = source.String
	if lastUsed.Valid {
		c.LastUsedAt = lastUsed.Time
	}
	return c, nil
}
