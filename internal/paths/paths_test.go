package paths_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/paths"
)

// setenv sets an env var and returns a cleanup function that restores the original value.
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

// TestDataDir_DefaultsUnderHome checks the default path when no XDG env is set.
func TestDataDir_DefaultsUnderHome(t *testing.T) {
	unsetenv(t, "XDG_DATA_HOME")

	got, err := paths.DataDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".local", "share", "recall")
	if got != want {
		t.Errorf("DataDir() = %q, want %q", got, want)
	}
}

// TestDataDir_RespectsXDG checks that XDG_DATA_HOME overrides the default.
func TestDataDir_RespectsXDG(t *testing.T) {
	setenv(t, "XDG_DATA_HOME", "/custom/data")

	got, err := paths.DataDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/custom/data/recall" {
		t.Errorf("DataDir() = %q, want /custom/data/recall", got)
	}
}

// TestConfigDir_DefaultsUnderHome checks the default config path.
func TestConfigDir_DefaultsUnderHome(t *testing.T) {
	unsetenv(t, "XDG_CONFIG_HOME")

	got, err := paths.ConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "recall")
	if got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

// TestConfigDir_RespectsXDG checks that XDG_CONFIG_HOME overrides the default.
func TestConfigDir_RespectsXDG(t *testing.T) {
	setenv(t, "XDG_CONFIG_HOME", "/custom/config")

	got, err := paths.ConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/custom/config/recall" {
		t.Errorf("ConfigDir() = %q, want /custom/config/recall", got)
	}
}

// TestDBPath_DefaultInsideDataDir checks that the DB lives inside DataDir by default.
func TestDBPath_DefaultInsideDataDir(t *testing.T) {
	unsetenv(t, "RECALL_DB_PATH")
	unsetenv(t, "XDG_DATA_HOME")

	got, err := paths.DBPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasSuffix(got, "recall.db") {
		t.Errorf("DBPath() = %q, expected suffix recall.db", got)
	}

	dataDir, _ := paths.DataDir()
	if filepath.Dir(got) != dataDir {
		t.Errorf("DBPath() dir = %q, want DataDir() %q", filepath.Dir(got), dataDir)
	}
}

// TestDBPath_EnvOverride checks that RECALL_DB_PATH completely overrides the default.
func TestDBPath_EnvOverride(t *testing.T) {
	setenv(t, "RECALL_DB_PATH", "/tmp/my-test.db")

	got, err := paths.DBPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/tmp/my-test.db" {
		t.Errorf("DBPath() = %q, want /tmp/my-test.db", got)
	}
}

// TestDBPath_XDGPropagates checks that XDG_DATA_HOME flows through to DBPath.
func TestDBPath_XDGPropagates(t *testing.T) {
	unsetenv(t, "RECALL_DB_PATH")
	setenv(t, "XDG_DATA_HOME", "/xdg/data")

	got, err := paths.DBPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/xdg/data/recall/recall.db"
	if got != want {
		t.Errorf("DBPath() = %q, want %q", got, want)
	}
}

// TestConfigPath_InsideConfigDir checks that ConfigPath lives inside ConfigDir.
func TestConfigPath_InsideConfigDir(t *testing.T) {
	unsetenv(t, "XDG_CONFIG_HOME")

	got, err := paths.ConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasSuffix(got, "config.yaml") {
		t.Errorf("ConfigPath() = %q, expected suffix config.yaml", got)
	}

	configDir, _ := paths.ConfigDir()
	if filepath.Dir(got) != configDir {
		t.Errorf("ConfigPath() dir = %q, want ConfigDir() %q", filepath.Dir(got), configDir)
	}
}

// TestConfigPath_XDGPropagates checks that XDG_CONFIG_HOME flows through to ConfigPath.
func TestConfigPath_XDGPropagates(t *testing.T) {
	setenv(t, "XDG_CONFIG_HOME", "/xdg/config")

	got, err := paths.ConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/xdg/config/recall/config.yaml"
	if got != want {
		t.Errorf("ConfigPath() = %q, want %q", got, want)
	}
}

// TestSourcesDir_InsideDataDir checks that SourcesDir lives inside DataDir.
func TestSourcesDir_InsideDataDir(t *testing.T) {
	unsetenv(t, "XDG_DATA_HOME")

	got, err := paths.SourcesDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dataDir, _ := paths.DataDir()
	want := filepath.Join(dataDir, "sources")
	if got != want {
		t.Errorf("SourcesDir() = %q, want %q", got, want)
	}
}

// TestSourcesDir_XDGPropagates checks that XDG_DATA_HOME flows through to SourcesDir.
func TestSourcesDir_XDGPropagates(t *testing.T) {
	setenv(t, "XDG_DATA_HOME", "/xdg/data")

	got, err := paths.SourcesDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/xdg/data/recall/sources"
	if got != want {
		t.Errorf("SourcesDir() = %q, want %q", got, want)
	}
}
