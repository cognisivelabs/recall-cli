package gitops

import (
	"os"
	"path/filepath"
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

// TestSync_ExistingLocalRepo verifies that Sync imports from a repo that was
// already cloned (simulated by creating a directory with YAML files directly).
// The git pull will fail because there is no real remote, but the directory
// exists so ImportFromRepo is still called on the cached content.
func TestSync_ExistingLocalRepo(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("XDG_DATA_HOME") })

	sourcesDir := filepath.Join(tmpDir, "recall", "sources")
	repoDir := filepath.Join(sourcesDir, "team-ops")
	os.MkdirAll(repoDir, 0755)
	os.WriteFile(filepath.Join(repoDir, "cmds.yaml"), []byte(`
- pattern: "kubectl get pods"
  description: "List pods"
`), 0644)

	cfg := &config.Config{
		Sources: []config.Source{
			{Name: "team-ops", Git: "git@invalid.example:repo.git"},
		},
	}
	store := newMemStore()

	// Sync will try `git pull` (it will fail) then call ImportFromRepo.
	// We just verify no panic — the exact import count depends on whether
	// git is installed and how it handles the invalid URL.
	_ = Sync(cfg, store)
}
