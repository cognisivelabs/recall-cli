package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isolateCmdEnv redirects all path-related env vars to a temp dir so
// config commands don't touch the real ~/.config/recall directory.
func isolateCmdEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	for _, key := range []string{"HOME", "XDG_DATA_HOME", "XDG_CONFIG_HOME", "RECALL_DB_PATH"} {
		prev, existed := os.LookupEnv(key)
		os.Unsetenv(key)
		t.Cleanup(func() {
			if existed {
				os.Setenv(key, prev)
			} else {
				os.Unsetenv(key)
			}
		})
	}
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	return tmpDir
}

// TestConfigPathCmd_PrintsPath verifies `recall config path` prints a path ending in config.yaml.
func TestConfigPathCmd_PrintsPath(t *testing.T) {
	isolateCmdEnv(t)

	cmd := NewConfigCmd()
	stdout, _, err := execCmd(cmd, "path")
	if err != nil {
		t.Fatalf("config path: %v", err)
	}
	if !strings.Contains(stdout, "config.yaml") {
		t.Errorf("expected config.yaml in output, got %q", stdout)
	}
}

// TestConfigInitCmd_CreatesFile verifies `recall config init` creates a config file.
func TestConfigInitCmd_CreatesFile(t *testing.T) {
	tmpDir := isolateCmdEnv(t)

	cmd := NewConfigCmd()
	_, _, err := execCmd(cmd, "init")
	if err != nil {
		t.Fatalf("config init: %v", err)
	}

	expected := filepath.Join(tmpDir, ".config", "recall", "config.yaml")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Errorf("config file not created at %s", expected)
	}
}

// TestConfigInitCmd_ErrorsIfAlreadyExists verifies a second `config init` returns an error.
func TestConfigInitCmd_ErrorsIfAlreadyExists(t *testing.T) {
	isolateCmdEnv(t)

	cmd := NewConfigCmd()
	if _, _, err := execCmd(cmd, "init"); err != nil {
		t.Fatalf("first config init: %v", err)
	}

	cmd2 := NewConfigCmd()
	_, _, err := execCmd(cmd2, "init")
	if err == nil {
		t.Fatal("expected error on second config init, got nil")
	}
}

// TestConfigShowCmd_PrintsTheme verifies `recall config show` prints the theme.
func TestConfigShowCmd_PrintsTheme(t *testing.T) {
	isolateCmdEnv(t)

	cmd := NewConfigCmd()
	stdout, _, err := execCmd(cmd, "show")
	if err != nil {
		t.Fatalf("config show: %v", err)
	}
	if !strings.Contains(stdout, "Theme:") {
		t.Errorf("expected 'Theme:' in output, got %q", stdout)
	}
}

// TestConfigShowCmd_ListsSources verifies `recall config show` lists sources.
func TestConfigShowCmd_ListsSources(t *testing.T) {
	isolateCmdEnv(t)

	cmd := NewConfigCmd()
	stdout, _, err := execCmd(cmd, "show")
	if err != nil {
		t.Fatalf("config show: %v", err)
	}
	if !strings.Contains(stdout, "Sources:") {
		t.Errorf("expected 'Sources:' in output, got %q", stdout)
	}
	// Default config has a "personal" source.
	if !strings.Contains(stdout, "personal") {
		t.Errorf("expected default 'personal' source in output, got %q", stdout)
	}
}

// TestConfigShowCmd_GitSourceDisplayed verifies git sources are shown with their URL.
func TestConfigShowCmd_GitSourceDisplayed(t *testing.T) {
	tmpDir := isolateCmdEnv(t)

	// Write a config that includes a git source.
	cfgDir := filepath.Join(tmpDir, ".config", "recall")
	os.MkdirAll(cfgDir, 0755)
	os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(`
theme: default
sources:
  - name: team-ops
    git: git@github.com:org/ops.git
`), 0644)

	cmd := NewConfigCmd()
	stdout, _, err := execCmd(cmd, "show")
	if err != nil {
		t.Fatalf("config show: %v", err)
	}
	if !strings.Contains(stdout, "git@github.com:org/ops.git") {
		t.Errorf("expected git URL in output, got %q", stdout)
	}
}
