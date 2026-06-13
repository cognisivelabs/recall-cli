//go:build !windows

package paths

import (
	"os"
	"path/filepath"
)

// defaultDataDir returns the XDG-compliant data directory: ~/.local/share/recall.
func defaultDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", appName), nil
}

// defaultConfigDir returns the XDG-compliant config directory: ~/.config/recall.
func defaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", appName), nil
}
