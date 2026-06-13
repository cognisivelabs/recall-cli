package gitops

import (
	"os"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/config"
)

// TestSync_NoGitSources verifies Sync is a no-op when only local sources are configured.
func TestSync_NoGitSources(t *testing.T) {
	cfg := &config.Config{
		Sources: []config.Source{
			{Name: "personal", Path: "/tmp/recall.db"},
		},
	}
	store := newMemStore()

	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("XDG_DATA_HOME") })

	if err := Sync(cfg, store); err != nil {
		t.Fatalf("Sync with no git sources: %v", err)
	}
	if len(store.commands) != 0 {
		t.Errorf("expected 0 imports for local-only config, got %d", len(store.commands))
	}
}

// TestSync_EmptyConfig verifies Sync handles a config with no sources.
func TestSync_EmptyConfig(t *testing.T) {
	cfg := &config.Config{}
	store := newMemStore()

	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("XDG_DATA_HOME") })

	if err := Sync(cfg, store); err != nil {
		t.Fatalf("Sync with empty config: %v", err)
	}
}
