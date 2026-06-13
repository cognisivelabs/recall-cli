package storage

import "time"

// Command is the core domain type: a shell command pattern with metadata.
// Pattern is the unique key — it may contain {{placeholder}} tokens.
// Tags is a comma-separated string (e.g. "k8s,debug").
// WorkspaceFilter is a path glob used to surface the command when the user is
// in a matching directory.
type Command struct {
	ID              int       `json:"id"`
	Pattern         string    `json:"pattern"`
	Description     string    `json:"description"`
	Tags            string    `json:"tags"` // Comma-separated or JSON
	WorkspaceFilter string    `json:"workspace_filter"`
	Source          string    `json:"source"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	LastUsedAt      time.Time `json:"last_used_at"`
	UsageCount      int       `json:"usage_count"`
}

// Storage is the persistence interface for recall commands.
// All callers (commands, TUI) depend on this interface so the underlying
// database can be swapped or mocked in tests.
type Storage interface {
	List() ([]Command, error)
	GetByID(id int) (*Command, error)
	GetByPattern(pattern string) (*Command, error)
	Delete(id int) error
	Update(c Command) error
	Upsert(c Command) error
	RecordUsage(id int) error
	Close() error
}
