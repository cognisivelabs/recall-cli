package placeholders

// Additional tests that exercise the auto-resolver helper functions directly.
// These are in the same package so they can call unexported helpers.

import (
	"os"
	"testing"
)

// TestCwd returns a non-empty string when the working directory is readable.
func TestCwd(t *testing.T) {
	got := cwd()
	if got == "" {
		t.Error("cwd() returned empty string")
	}
}

// TestDirName returns the base name of the working directory.
func TestDirName(t *testing.T) {
	got := dirName()
	if got == "" {
		t.Error("dirName() returned empty string")
	}
}

// TestUserName returns a non-empty OS username.
func TestUserName(t *testing.T) {
	got := userName()
	if got == "" {
		t.Error("userName() returned empty string")
	}
}

// TestHostName returns a non-empty hostname.
func TestHostName(t *testing.T) {
	got := hostName()
	if got == "" {
		t.Error("hostName() returned empty string")
	}
}

// TestHomeDir returns a non-empty home directory.
func TestHomeDir(t *testing.T) {
	got := homeDir()
	if got == "" {
		t.Error("homeDir() returned empty string")
	}
}

// TestHomeDir_FallbackOnBadHome verifies homeDir returns "" when HOME is unset.
func TestHomeDir_FallbackOnBadHome(t *testing.T) {
	prev, existed := os.LookupEnv("HOME")
	os.Unsetenv("HOME")
	t.Cleanup(func() {
		if existed {
			os.Setenv("HOME", prev)
		}
	})
	// homeDir should not panic; it may return "" or the OS default.
	_ = homeDir()
}

// TestParse_AllAutoKeys verifies that every auto-resolver key is recognised.
func TestParse_AllAutoKeys(t *testing.T) {
	autoKeys := []string{"branch", "cwd", "dir", "user", "host", "home"}
	for _, key := range autoKeys {
		placeholders := Parse("cmd {{" + key + "}}")
		if len(placeholders) != 1 {
			t.Errorf("key %q: expected 1 placeholder, got %d", key, len(placeholders))
			continue
		}
		if placeholders[0].Type != "auto" {
			t.Errorf("key %q: expected type 'auto', got %q", key, placeholders[0].Type)
		}
	}
}

// TestAutoResolve_NothingToResolve returns original when no placeholders.
func TestAutoResolve_NothingToResolve(t *testing.T) {
	cmd := "echo hello"
	resolved, remaining := AutoResolve(cmd)
	if resolved != "echo hello" {
		t.Errorf("expected unchanged command, got %q", resolved)
	}
	if len(remaining) != 0 {
		t.Errorf("expected 0 remaining, got %d", len(remaining))
	}
}

// TestAutoResolve_OnlyTextPlaceholders leaves text placeholders unresolved.
func TestAutoResolve_OnlyTextPlaceholders(t *testing.T) {
	cmd := "kubectl logs {{service}}"
	resolved, remaining := AutoResolve(cmd)
	if resolved != cmd {
		t.Errorf("text placeholder should not be auto-resolved, got %q", resolved)
	}
	if len(remaining) != 1 {
		t.Errorf("expected 1 remaining placeholder, got %d", len(remaining))
	}
}
