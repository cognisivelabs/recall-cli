package main

import (
	"os"
	"strings"
	"testing"
)

// TestInitCmd_Zsh verifies `recall init zsh` outputs a non-empty shell script.
func TestInitCmd_Zsh(t *testing.T) {
	cmd := NewInitCmd()
	stdout, _, err := execCmd(cmd, "zsh")
	if err != nil {
		t.Fatalf("init zsh: %v", err)
	}
	if !strings.Contains(stdout, "recall") {
		t.Errorf("expected recall reference in zsh init script, got %q", stdout)
	}
}

// TestInitCmd_Bash verifies `recall init bash` outputs a non-empty shell script.
func TestInitCmd_Bash(t *testing.T) {
	cmd := NewInitCmd()
	stdout, _, err := execCmd(cmd, "bash")
	if err != nil {
		t.Fatalf("init bash: %v", err)
	}
	if stdout == "" {
		t.Error("expected non-empty bash init script")
	}
}

// TestInitCmd_AutoDetectFromSHELL verifies `recall init` with no args uses $SHELL.
func TestInitCmd_AutoDetectFromSHELL(t *testing.T) {
	prev, existed := os.LookupEnv("SHELL")
	os.Setenv("SHELL", "/bin/zsh")
	t.Cleanup(func() {
		if existed {
			os.Setenv("SHELL", prev)
		} else {
			os.Unsetenv("SHELL")
		}
	})

	cmd := NewInitCmd()
	stdout, _, err := execCmd(cmd)
	if err != nil {
		t.Fatalf("init (auto-detect): %v", err)
	}
	if stdout == "" {
		t.Error("expected non-empty init script when SHELL is set")
	}
}

// TestInitCmd_UnsupportedShellErrors verifies that an unknown shell returns an error.
func TestInitCmd_UnsupportedShellErrors(t *testing.T) {
	cmd := NewInitCmd()
	_, _, err := execCmd(cmd, "fish")
	if err == nil {
		t.Fatal("expected error for unsupported shell 'fish', got nil")
	}
}
