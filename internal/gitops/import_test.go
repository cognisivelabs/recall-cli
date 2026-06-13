package gitops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
)

// memStore is a minimal in-memory Storage for import tests — no SQLite needed.
type memStore struct {
	commands map[string]storage.Command
}

func newMemStore() *memStore {
	return &memStore{commands: make(map[string]storage.Command)}
}

func (m *memStore) Upsert(c storage.Command) error {
	m.commands[c.Pattern] = c
	return nil
}

func (m *memStore) List() ([]storage.Command, error) {
	out := make([]storage.Command, 0, len(m.commands))
	for _, c := range m.commands {
		out = append(out, c)
	}
	return out, nil
}

func (m *memStore) GetByID(_ int) (*storage.Command, error)        { return nil, nil }
func (m *memStore) GetByPattern(_ string) (*storage.Command, error) { return nil, nil }
func (m *memStore) Delete(_ int) error                              { return nil }
func (m *memStore) Update(_ storage.Command) error                  { return nil }
func (m *memStore) RecordUsage(_ int) error                         { return nil }
func (m *memStore) Close() error                                    { return nil }

// --- ImportFromRepo ---

func TestImportFromRepo_StructuredYAML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "ops.yaml"), []byte(`
commands:
  - pattern: "kubectl get pods"
    description: "List pods"
    tags: [k8s, debug]
  - pattern: "docker ps"
    description: "Containers"
    tags: [docker]
`), 0644)

	store := newMemStore()
	n, err := ImportFromRepo(store, dir, "team-ops")
	if err != nil {
		t.Fatalf("ImportFromRepo: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 imported, got %d", n)
	}
	if c, ok := store.commands["kubectl get pods"]; !ok {
		t.Error("kubectl get pods not imported")
	} else if c.Source != "team-ops" {
		t.Errorf("expected source 'team-ops', got %q", c.Source)
	}
}

func TestImportFromRepo_FlatYAML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "cmds.yaml"), []byte(`
- pattern: "echo hello"
  description: "Hello"
`), 0644)

	store := newMemStore()
	n, err := ImportFromRepo(store, dir, "personal")
	if err != nil {
		t.Fatalf("ImportFromRepo: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 imported, got %d", n)
	}
}

func TestImportFromRepo_SkipsEmptyPattern(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "cmds.yaml"), []byte(`
- pattern: ""
  description: "no pattern"
- pattern: "echo hi"
  description: "valid"
`), 0644)

	store := newMemStore()
	n, err := ImportFromRepo(store, dir, "src")
	if err != nil {
		t.Fatalf("ImportFromRepo: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 imported (empty pattern skipped), got %d", n)
	}
}

func TestImportFromRepo_TagsJoinedWithComma(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "cmds.yaml"), []byte(`
commands:
  - pattern: "kubectl get pods"
    tags: [k8s, debug, prod]
`), 0644)

	store := newMemStore()
	ImportFromRepo(store, dir, "src")

	c, ok := store.commands["kubectl get pods"]
	if !ok {
		t.Fatal("command not imported")
	}
	if !strings.Contains(c.Tags, "k8s") || !strings.Contains(c.Tags, "debug") {
		t.Errorf("expected comma-joined tags, got %q", c.Tags)
	}
}

func TestImportFromRepo_SkipsMalformedFile(t *testing.T) {
	dir := t.TempDir()
	// bad file — should be skipped, not fail the whole import
	os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("not: [valid: yaml: {{"), 0644)
	os.WriteFile(filepath.Join(dir, "good.yaml"), []byte(`
- pattern: "echo ok"
`), 0644)

	store := newMemStore()
	n, err := ImportFromRepo(store, dir, "src")
	if err != nil {
		t.Fatalf("ImportFromRepo should not fail on malformed file: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 imported from good file, got %d", n)
	}
}

func TestImportFromRepo_NestedDirectories(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "ops", "k8s")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(sub, "pods.yaml"), []byte(`
- pattern: "kubectl get pods"
`), 0644)

	store := newMemStore()
	n, err := ImportFromRepo(store, dir, "src")
	if err != nil {
		t.Fatalf("ImportFromRepo: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 imported from nested dir, got %d", n)
	}
}

func TestImportFromRepo_GitDirSkipped(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	os.MkdirAll(gitDir, 0755)
	// A yaml file inside .git should be ignored.
	os.WriteFile(filepath.Join(gitDir, "config.yaml"), []byte(`
- pattern: "should not be imported"
`), 0644)

	store := newMemStore()
	n, err := ImportFromRepo(store, dir, "src")
	if err != nil {
		t.Fatalf("ImportFromRepo: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 imported (.git dir should be skipped), got %d", n)
	}
}

func TestImportFromRepo_EmptyRepo(t *testing.T) {
	dir := t.TempDir()
	store := newMemStore()
	n, err := ImportFromRepo(store, dir, "src")
	if err != nil {
		t.Fatalf("ImportFromRepo on empty dir: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 imported from empty repo, got %d", n)
	}
}

func TestImportFromRepo_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.yaml"), []byte(`- pattern: "cmd a"`), 0644)
	os.WriteFile(filepath.Join(dir, "b.yml"), []byte(`- pattern: "cmd b"`), 0644)

	store := newMemStore()
	n, err := ImportFromRepo(store, dir, "src")
	if err != nil {
		t.Fatalf("ImportFromRepo: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 imported from 2 files, got %d", n)
	}
}
