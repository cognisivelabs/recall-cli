//go:build windows

package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

// defaultDataDir returns the Windows-native data directory: %APPDATA%\recall.
// Both data and config share the same %APPDATA%\recall directory on Windows,
// matching the convention used by most Windows applications.
func defaultDataDir() (string, error) {
	appdata := os.Getenv("APPDATA")
	if appdata == "" {
		return "", fmt.Errorf("%%APPDATA%% is not set; set XDG_DATA_HOME or RECALL_DB_PATH to specify a location")
	}
	return filepath.Join(appdata, appName), nil
}

// defaultConfigDir returns the Windows-native config directory: %APPDATA%\recall.
// On Windows, config lives alongside data in %APPDATA%\recall rather than in a
// separate directory, following the Windows single-folder convention.
func defaultConfigDir() (string, error) {
	return defaultDataDir()
}
