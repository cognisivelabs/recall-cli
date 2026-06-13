// Package paths resolves all filesystem locations used by recall.
//
// Every path is derived from environment variables so the app is fully
// relocatable without recompiling (12-factor III).
//
// Resolution order (same on all platforms):
//
//  1. Explicit overrides
//     RECALL_DB_PATH   → overrides the SQLite file path entirely
//     XDG_DATA_HOME    → overrides the data directory root
//     XDG_CONFIG_HOME  → overrides the config directory root
//
//  2. Platform defaults (when no override is set)
//     Windows → %APPDATA%\recall  (e.g. C:\Users\Alice\AppData\Roaming\recall)
//     Unix    → data:   ~/.local/share/recall
//               config: ~/.config/recall
//
// The platform default logic lives in paths_windows.go / paths_unix.go so
// this file stays free of runtime.GOOS checks.
package paths

import (
	"os"
	"path/filepath"
)

const appName = "recall"

// DataDir returns the directory that holds recall's data files (SQLite DB,
// cloned git sources). XDG_DATA_HOME overrides the platform default.
func DataDir() (string, error) {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, appName), nil
	}
	return defaultDataDir()
}

// ConfigDir returns the directory that holds recall's config file.
// XDG_CONFIG_HOME overrides the platform default.
func ConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appName), nil
	}
	return defaultConfigDir()
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

// ConfigPath returns the full path to the YAML config file (inside ConfigDir).
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
