// Package paths resolves all filesystem locations used by recall.
//
// Every path is derived from environment variables so the app is fully
// relocatable without recompiling (12-factor III).
//
// Resolution order for data and config directories:
//
//  1. XDG_DATA_HOME / XDG_CONFIG_HOME — explicit override, works on all platforms
//  2. %APPDATA%\recall               — Windows default
//  3. ~/.local/share/recall          — Linux/macOS data default
//     ~/.config/recall               — Linux/macOS config default
package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const appName = "recall"

// DataDir returns the directory that holds recall's data files (SQLite DB,
// cloned git sources).
//
// Windows: %APPDATA%\recall  (e.g. C:\Users\Alice\AppData\Roaming\recall)
// Unix:    ~/.local/share/recall
// Override: set XDG_DATA_HOME to any path on any platform.
func DataDir() (string, error) {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, appName), nil
	}
	if runtime.GOOS == "windows" {
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			return "", fmt.Errorf("%%APPDATA%% is not set; set XDG_DATA_HOME or RECALL_DB_PATH to specify a location")
		}
		return filepath.Join(appdata, appName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", appName), nil
}

// ConfigDir returns the directory that holds recall's config file.
//
// Windows: %APPDATA%\recall  (same directory as data, matching Windows conventions)
// Unix:    ~/.config/recall
// Override: set XDG_CONFIG_HOME to any path on any platform.
func ConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appName), nil
	}
	if runtime.GOOS == "windows" {
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			return "", fmt.Errorf("%%APPDATA%% is not set; set XDG_CONFIG_HOME to specify a location")
		}
		return filepath.Join(appdata, appName), nil
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
