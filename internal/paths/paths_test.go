package paths_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/paths"
)

// setenv sets an env var for the duration of the test and restores it after.
func setenv(t *testing.T, key, value string) {
	t.Helper()
	prev, existed := os.LookupEnv(key)
	os.Setenv(key, value)
	t.Cleanup(func() {
		if existed {
			os.Setenv(key, prev)
		} else {
			os.Unsetenv(key)
		}
	})
}

// unsetenv removes an env var for the duration of the test and restores it after.
func unsetenv(t *testing.T, key string) {
	t.Helper()
	prev, existed := os.LookupEnv(key)
	os.Unsetenv(key)
	t.Cleanup(func() {
		if existed {
			os.Setenv(key, prev)
		}
	})
}

// --- DataDir ---

// TestDataDir_PlatformDefault verifies the no-override path is correct for the
// current platform: %APPDATA%\recall on Windows, ~/.local/share/recall on Unix.
func TestDataDir_PlatformDefault(t *testing.T) {
	unsetenv(t, "XDG_DATA_HOME")

	got, err := paths.DataDir()
	if err != nil {
		t.Fatalf("DataDir(): %v", err)
	}

	if runtime.GOOS == "windows" {
		appdata := os.Getenv("APPDATA")
		want := filepath.Join(appdata, "recall")
		if got != want {
			t.Errorf("DataDir() = %q, want %q", got, want)
		}
	} else {
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".local", "share", "recall")
		if got != want {
			t.Errorf("DataDir() = %q, want %q", got, want)
		}
	}
}

// TestDataDir_XDGOverride verifies XDG_DATA_HOME takes priority on all platforms.
func TestDataDir_XDGOverride(t *testing.T) {
	xdg := t.TempDir()
	setenv(t, "XDG_DATA_HOME", xdg)

	got, err := paths.DataDir()
	if err != nil {
		t.Fatalf("DataDir(): %v", err)
	}
	want := filepath.Join(xdg, "recall")
	if got != want {
		t.Errorf("DataDir() = %q, want %q", got, want)
	}
}

// TestDataDir_APPDATAIgnoredOnUnix verifies that APPDATA has no effect on Unix.
func TestDataDir_APPDATAIgnoredOnUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-only test")
	}
	unsetenv(t, "XDG_DATA_HOME")
	setenv(t, "APPDATA", t.TempDir())

	got, err := paths.DataDir()
	if err != nil {
		t.Fatalf("DataDir(): %v", err)
	}
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".local", "share", "recall")
	if got != want {
		t.Errorf("APPDATA should be ignored on Unix; DataDir() = %q, want %q", got, want)
	}
}

// TestDataDir_WindowsAPPDATAMissing verifies an error is returned when APPDATA is unset on Windows.
func TestDataDir_WindowsAPPDATAMissing(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}
	unsetenv(t, "XDG_DATA_HOME")
	unsetenv(t, "APPDATA")

	_, err := paths.DataDir()
	if err == nil {
		t.Fatal("expected error when APPDATA is unset on Windows, got nil")
	}
}

// --- ConfigDir ---

// TestConfigDir_PlatformDefault verifies the no-override path is correct for the
// current platform: %APPDATA%\recall on Windows, ~/.config/recall on Unix.
func TestConfigDir_PlatformDefault(t *testing.T) {
	unsetenv(t, "XDG_CONFIG_HOME")

	got, err := paths.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir(): %v", err)
	}

	if runtime.GOOS == "windows" {
		appdata := os.Getenv("APPDATA")
		want := filepath.Join(appdata, "recall")
		if got != want {
			t.Errorf("ConfigDir() = %q, want %q", got, want)
		}
	} else {
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".config", "recall")
		if got != want {
			t.Errorf("ConfigDir() = %q, want %q", got, want)
		}
	}
}

// TestConfigDir_XDGOverride verifies XDG_CONFIG_HOME takes priority on all platforms.
func TestConfigDir_XDGOverride(t *testing.T) {
	xdg := t.TempDir()
	setenv(t, "XDG_CONFIG_HOME", xdg)

	got, err := paths.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir(): %v", err)
	}
	want := filepath.Join(xdg, "recall")
	if got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

// TestDataConfigDir_SameOnWindows verifies data and config share %APPDATA%\recall on Windows.
func TestDataConfigDir_SameOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}
	unsetenv(t, "XDG_DATA_HOME")
	unsetenv(t, "XDG_CONFIG_HOME")

	dataDir, err := paths.DataDir()
	if err != nil {
		t.Fatalf("DataDir(): %v", err)
	}
	configDir, err := paths.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir(): %v", err)
	}
	if dataDir != configDir {
		t.Errorf("on Windows DataDir and ConfigDir should be equal: data=%q config=%q", dataDir, configDir)
	}
}

// --- DBPath ---

// TestDBPath_InsideDataDir verifies the DB lives inside DataDir by default.
func TestDBPath_InsideDataDir(t *testing.T) {
	unsetenv(t, "RECALL_DB_PATH")
	unsetenv(t, "XDG_DATA_HOME")

	got, err := paths.DBPath()
	if err != nil {
		t.Fatalf("DBPath(): %v", err)
	}
	if !strings.HasSuffix(got, "recall.db") {
		t.Errorf("DBPath() = %q, expected suffix recall.db", got)
	}
	dataDir, _ := paths.DataDir()
	if filepath.Dir(got) != dataDir {
		t.Errorf("DBPath() dir = %q, want DataDir() %q", filepath.Dir(got), dataDir)
	}
}

// TestDBPath_EnvOverride verifies RECALL_DB_PATH overrides everything.
func TestDBPath_EnvOverride(t *testing.T) {
	override := filepath.Join(t.TempDir(), "my-test.db")
	setenv(t, "RECALL_DB_PATH", override)

	got, err := paths.DBPath()
	if err != nil {
		t.Fatalf("DBPath(): %v", err)
	}
	if got != override {
		t.Errorf("DBPath() = %q, want %q", got, override)
	}
}

// TestDBPath_XDGPropagates verifies XDG_DATA_HOME flows through to DBPath.
func TestDBPath_XDGPropagates(t *testing.T) {
	xdg := t.TempDir()
	unsetenv(t, "RECALL_DB_PATH")
	setenv(t, "XDG_DATA_HOME", xdg)

	got, err := paths.DBPath()
	if err != nil {
		t.Fatalf("DBPath(): %v", err)
	}
	want := filepath.Join(xdg, "recall", "recall.db")
	if got != want {
		t.Errorf("DBPath() = %q, want %q", got, want)
	}
}

// --- ConfigPath ---

// TestConfigPath_InsideConfigDir verifies ConfigPath lives inside ConfigDir.
func TestConfigPath_InsideConfigDir(t *testing.T) {
	unsetenv(t, "XDG_CONFIG_HOME")

	got, err := paths.ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath(): %v", err)
	}
	if !strings.HasSuffix(got, "config.yaml") {
		t.Errorf("ConfigPath() = %q, expected suffix config.yaml", got)
	}
	configDir, _ := paths.ConfigDir()
	if filepath.Dir(got) != configDir {
		t.Errorf("ConfigPath() dir = %q, want ConfigDir() %q", filepath.Dir(got), configDir)
	}
}

// TestConfigPath_XDGPropagates verifies XDG_CONFIG_HOME flows through to ConfigPath.
func TestConfigPath_XDGPropagates(t *testing.T) {
	xdg := t.TempDir()
	setenv(t, "XDG_CONFIG_HOME", xdg)

	got, err := paths.ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath(): %v", err)
	}
	want := filepath.Join(xdg, "recall", "config.yaml")
	if got != want {
		t.Errorf("ConfigPath() = %q, want %q", got, want)
	}
}

// --- SourcesDir ---

// TestSourcesDir_InsideDataDir verifies SourcesDir lives inside DataDir.
func TestSourcesDir_InsideDataDir(t *testing.T) {
	unsetenv(t, "XDG_DATA_HOME")

	got, err := paths.SourcesDir()
	if err != nil {
		t.Fatalf("SourcesDir(): %v", err)
	}
	dataDir, _ := paths.DataDir()
	want := filepath.Join(dataDir, "sources")
	if got != want {
		t.Errorf("SourcesDir() = %q, want %q", got, want)
	}
}

// TestSourcesDir_XDGPropagates verifies XDG_DATA_HOME flows through to SourcesDir.
func TestSourcesDir_XDGPropagates(t *testing.T) {
	xdg := t.TempDir()
	setenv(t, "XDG_DATA_HOME", xdg)

	got, err := paths.SourcesDir()
	if err != nil {
		t.Fatalf("SourcesDir(): %v", err)
	}
	want := filepath.Join(xdg, "recall", "sources")
	if got != want {
		t.Errorf("SourcesDir() = %q, want %q", got, want)
	}
}
