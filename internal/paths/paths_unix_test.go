//go:build !windows

package paths_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/paths"
)

// TestDataDir_UnixDefault verifies the Unix fallback is ~/.local/share/recall.
func TestDataDir_UnixDefault(t *testing.T) {
	unsetenv(t, "XDG_DATA_HOME")

	got, err := paths.DataDir()
	if err != nil {
		t.Fatalf("DataDir(): %v", err)
	}

	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".local", "share", "recall")
	if got != want {
		t.Errorf("DataDir() = %q, want %q", got, want)
	}
}

// TestConfigDir_UnixDefault verifies the Unix fallback is ~/.config/recall.
func TestConfigDir_UnixDefault(t *testing.T) {
	unsetenv(t, "XDG_CONFIG_HOME")

	got, err := paths.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir(): %v", err)
	}

	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "recall")
	if got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

// TestDataDir_APPDATAIgnoredOnUnix verifies that APPDATA has no effect on Unix.
func TestDataDir_APPDATAIgnoredOnUnix(t *testing.T) {
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
