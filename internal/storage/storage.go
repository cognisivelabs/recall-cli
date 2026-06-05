package storage

import "time"

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
