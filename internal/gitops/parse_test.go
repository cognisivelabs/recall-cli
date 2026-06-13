package gitops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
)

// memStore is a minimal in-memory Storage used across gitops tests — no SQLite needed.
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

func (m *memStore) GetByID(_ int) (*storage.Command, error)         { return nil, nil }
func (m *memStore) GetByPattern(_ string) (*storage.Command, error) { return nil, nil }
func (m *memStore) Delete(_ int) error                              { return nil }
func (m *memStore) Update(_ storage.Command) error                  { return nil }
func (m *memStore) RecordUsage(_ int) error                         { return nil }
func (m *memStore) Close() error                                    { return nil }

// Compile-time check that memStore satisfies storage.Storage.
var _ storage.Storage = (*memStore)(nil)

// --- parseCommandFile ---

func TestParseCommandFile_StructuredFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "commands.yaml")
	os.WriteFile(path, []byte(`
commands:
  - pattern: "kubectl get pods"
    description: "List pods"
    tags: [k8s, debug]
  - pattern: "docker ps"
    description: "List containers"
    tags: [docker]
`), 0644)

	entries, err := parseCommandFile(path)
	if err != nil {
		t.Fatalf("parseCommandFile: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Pattern != "kubectl get pods" {
		t.Errorf("expected 'kubectl get pods', got %q", entries[0].Pattern)
	}
	if len(entries[0].Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(entries[0].Tags))
	}
}

func TestParseCommandFile_FlatFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "commands.yaml")
	os.WriteFile(path, []byte(`
- pattern: "echo hello"
  description: "Hello world"
  tags: [test]
- pattern: "ls -la"
  description: "List files"
`), 0644)

	entries, err := parseCommandFile(path)
	if err != nil {
		t.Fatalf("parseCommandFile: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[1].Pattern != "ls -la" {
		t.Errorf("expected 'ls -la', got %q", entries[1].Pattern)
	}
}

func TestParseCommandFile_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	os.WriteFile(path, []byte("not: [valid: yaml: {{"), 0644)

	_, err := parseCommandFile(path)
	if err == nil {
		t.Fatal("expected error for invalid yaml")
	}
}

func TestParseCommandFile_EmptyPattern(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "commands.yaml")
	os.WriteFile(path, []byte(`
- pattern: ""
  description: "empty"
- pattern: "real command"
  description: "valid"
`), 0644)

	entries, err := parseCommandFile(path)
	if err != nil {
		t.Fatalf("parseCommandFile: %v", err)
	}
	// Filtering empty patterns is the importer's job; the parser returns both.
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

// --- findCommandFiles ---

func TestFindCommandFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "commands.yaml"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "extra.yml"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte(""), 0644)

	sub := filepath.Join(dir, "ops")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(sub, "runbooks.yaml"), []byte(""), 0644)

	// .git directory must be skipped.
	gitDir := filepath.Join(dir, ".git")
	os.MkdirAll(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "config.yaml"), []byte(""), 0644)

	files, err := findCommandFiles(dir)
	if err != nil {
		t.Fatalf("findCommandFiles: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 yaml files (excluding .git), got %d: %v", len(files), files)
	}
}

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
	c, ok := store.commands["kubectl get pods"]
	if !ok {
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
