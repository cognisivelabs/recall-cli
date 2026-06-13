//go:build windows

package paths_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/paths"
)

// TestDataDir_WindowsDefault verifies the Windows fallback is %APPDATA%\recall.
func TestDataDir_WindowsDefault(t *testing.T) {
	appdata := t.TempDir()
	unsetenv(t, "XDG_DATA_HOME")
	setenv(t, "APPDATA", appdata)

	got, err := paths.DataDir()
	if err != nil {
		t.Fatalf("DataDir(): %v", err)
	}

	want := filepath.Join(appdata, "recall")
	if got != want {
		t.Errorf("DataDir() = %q, want %q", got, want)
	}
}

// TestConfigDir_WindowsDefault verifies the Windows fallback is %APPDATA%\recall.
func TestConfigDir_WindowsDefault(t *testing.T) {
	appdata := t.TempDir()
	unsetenv(t, "XDG_CONFIG_HOME")
	setenv(t, "APPDATA", appdata)

	got, err := paths.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir(): %v", err)
	}

	want := filepath.Join(appdata, "recall")
	if got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

// TestDataDir_WindowsAPPDATAMissing verifies an error is returned when APPDATA is unset.
func TestDataDir_WindowsAPPDATAMissing(t *testing.T) {
	unsetenv(t, "XDG_DATA_HOME")
	unsetenv(t, "APPDATA")

	_, err := paths.DataDir()
	if err == nil {
		t.Fatal("expected error when APPDATA is unset on Windows, got nil")
	}
}

// TestDataDir_XDGOverridesAPPDATA verifies XDG_DATA_HOME takes priority over APPDATA on Windows.
func TestDataDir_XDGOverridesAPPDATA(t *testing.T) {
	xdg := t.TempDir()
	setenv(t, "XDG_DATA_HOME", xdg)
	setenv(t, "APPDATA", t.TempDir())

	got, err := paths.DataDir()
	if err != nil {
		t.Fatalf("DataDir(): %v", err)
	}

	want := filepath.Join(xdg, "recall")
	if got != want {
		t.Errorf("XDG_DATA_HOME should take priority; DataDir() = %q, want %q", got, want)
	}
}

// TestConfigDir_XDGOverridesAPPDATA verifies XDG_CONFIG_HOME takes priority over APPDATA on Windows.
func TestConfigDir_XDGOverridesAPPDATA(t *testing.T) {
	xdg := t.TempDir()
	setenv(t, "XDG_CONFIG_HOME", xdg)
	setenv(t, "APPDATA", t.TempDir())

	got, err := paths.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir(): %v", err)
	}

	want := filepath.Join(xdg, "recall")
	if got != want {
		t.Errorf("XDG_CONFIG_HOME should take priority; ConfigDir() = %q, want %q", got, want)
	}
}

// TestDBPath_WindowsDefault verifies the DB lands inside %APPDATA%\recall on Windows.
func TestDBPath_WindowsDefault(t *testing.T) {
	appdata := t.TempDir()
	unsetenv(t, "RECALL_DB_PATH")
	unsetenv(t, "XDG_DATA_HOME")
	setenv(t, "APPDATA", appdata)

	got, err := paths.DBPath()
	if err != nil {
		t.Fatalf("DBPath(): %v", err)
	}

	want := filepath.Join(appdata, "recall", "recall.db")
	if got != want {
		t.Errorf("DBPath() = %q, want %q", got, want)
	}
}

// TestConfigPath_WindowsDefault verifies config.yaml lands inside %APPDATA%\recall on Windows.
func TestConfigPath_WindowsDefault(t *testing.T) {
	appdata := t.TempDir()
	unsetenv(t, "XDG_CONFIG_HOME")
	setenv(t, "APPDATA", appdata)

	got, err := paths.ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath(): %v", err)
	}

	want := filepath.Join(appdata, "recall", "config.yaml")
	if got != want {
		t.Errorf("ConfigPath() = %q, want %q", got, want)
	}
}

// TestDataDir_ConfigDir_SameOnWindows verifies data and config share the same root directory.
func TestDataDir_ConfigDir_SameOnWindows(t *testing.T) {
	appdata := t.TempDir()
	unsetenv(t, "XDG_DATA_HOME")
	unsetenv(t, "XDG_CONFIG_HOME")
	setenv(t, "APPDATA", appdata)

	dataDir, err := paths.DataDir()
	if err != nil {
		t.Fatalf("DataDir(): %v", err)
	}
	configDir, err := paths.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir(): %v", err)
	}

	if dataDir != configDir {
		t.Errorf("on Windows, DataDir and ConfigDir should be the same: data=%q config=%q", dataDir, configDir)
	}
}

// TestConfigPath_WindowsInConfigDir verifies config.yaml is inside ConfigDir.
func TestConfigPath_WindowsInConfigDir(t *testing.T) {
	appdata := t.TempDir()
	unsetenv(t, "XDG_CONFIG_HOME")
	setenv(t, "APPDATA", appdata)

	configDir, _ := paths.ConfigDir()
	configPath, err := paths.ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath(): %v", err)
	}

	wantPath := filepath.Join(configDir, "config.yaml")
	if configPath != wantPath {
		t.Errorf("ConfigPath() = %q, want %q", configPath, wantPath)
	}
}

func TestSourcesDir_WindowsDefault(t *testing.T) {
	appdata := t.TempDir()
	unsetenv(t, "XDG_DATA_HOME")
	setenv(t, "APPDATA", appdata)

	got, err := paths.SourcesDir()
	if err != nil {
		t.Fatalf("SourcesDir(): %v", err)
	}

	want := filepath.Join(appdata, "recall", "sources")
	if got != want {
		t.Errorf("SourcesDir() = %q, want %q", got, want)
	}
}
