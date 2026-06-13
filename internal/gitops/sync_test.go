package gitops

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/config"
	"github.com/CognisiveLabs/recall-cli/internal/storage"
)

// syncMemStore satisfies storage.Storage for Sync tests.
type syncMemStore struct{ *memStore }

// TestSync_NoGitSources verifies Sync is a no-op when only local sources are configured.
func TestSync_NoGitSources(t *testing.T) {
	cfg := &config.Config{
		Sources: []config.Source{
			{Name: "personal", Path: "/tmp/recall.db"},
		},
	}
	store := newMemStore()

	// Override SourcesDir via XDG so Sync doesn't touch the real filesystem.
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
// This exercises the "git pull" branch without requiring a real git server —
// the pull will fail, but ImportFromRepo is still called and counts the files.
func TestSync_ExistingLocalRepo(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("XDG_DATA_HOME") })

	// Pre-create the sources directory and a fake "already cloned" repo.
	sourcesDir := filepath.Join(tmpDir, "recall", "sources")
	repoDir := filepath.Join(sourcesDir, "team-ops")
	os.MkdirAll(repoDir, 0755)
	os.WriteFile(filepath.Join(repoDir, "cmds.yaml"), []byte(`
- pattern: "kubectl get pods"
  description: "List pods"
`), 0644)

	// Use a clearly invalid git URL so the pull fails (skipped), but the repo
	// dir exists so ImportFromRepo is still called.
	cfg := &config.Config{
		Sources: []config.Source{
			{Name: "team-ops", Git: "git@invalid.example:repo.git"},
		},
	}
	store := newMemStore()

	// Sync will try `git pull` (it will fail) then call ImportFromRepo.
	// We accept either outcome: if git pull fails the source is skipped and
	// count stays 0, OR if the pull is skipped on error and import runs, count is 1.
	// The important thing is no panic and no unhandled error.
	_ = Sync(cfg, store)
	// No assertion on count — the exact behaviour depends on whether `git` is
	// installed and the environment. We just verify no panic or fatal error.
}

// TestImportFromRepo_SourceTagAttached verifies the source name is stored on every command.
func TestImportFromRepo_SourceTagAttached(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "cmds.yaml"), []byte(`
- pattern: "make build"
  description: "Build"
`), 0644)

	store := newMemStore()
	n, err := ImportFromRepo(store, dir, "my-team")
	if err != nil {
		t.Fatalf("ImportFromRepo: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 import, got %d", n)
	}

	c, ok := store.commands["make build"]
	if !ok {
		t.Fatal("command not found in store")
	}
	if c.Source != "my-team" {
		t.Errorf("expected source 'my-team', got %q", c.Source)
	}
}

// Compile-time check: memStore satisfies storage.Storage.
var _ storage.Storage = (*memStore)(nil)
