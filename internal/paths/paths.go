// Package paths resolves all filesystem locations used by recall.
//
// Every path is derived from environment variables so the app is fully
// relocatable without recompiling (12-factor III).
//
// Priority for each directory:
//
//	RECALL_DB_PATH        → direct override for the SQLite file
//	XDG_DATA_HOME         → data root (default: ~/.local/share)
//	XDG_CONFIG_HOME       → config root (default: ~/.config)
package paths

import (
	"os"
	"path/filepath"
)

const appName = "recall"

// DataDir returns the directory that holds recall's data files (SQLite DB,
// cloned git sources). Respects XDG_DATA_HOME; falls back to ~/.local/share/recall.
func DataDir() (string, error) {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, appName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", appName), nil
}

// ConfigDir returns the directory that holds recall's config file.
// Respects XDG_CONFIG_HOME; falls back to ~/.config/recall.
func ConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", appName), nil
}

// DBPath returns the path to the SQLite database file.
// RECALL_DB_PATH overrides everything; otherwise it is DataDir()/recall.db.
func DBPath() (string, error) {
	if override := os.Getenv("RECALL_DB_PATH"); override != "" {
		return override, nil
	}
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "recall.db"), nil
}

// ConfigPath returns the path to the YAML config file (inside ConfigDir).
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// SourcesDir returns the directory where synced git repos are cloned.
// It lives inside DataDir so all recall data stays in one place.
func SourcesDir() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sources"), nil
}
