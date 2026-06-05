package gitops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCommandFile_StructuredFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "commands.yaml")

	content := `
commands:
  - pattern: "kubectl get pods"
    description: "List pods"
    tags: [k8s, debug]
  - pattern: "docker ps"
    description: "List containers"
    tags: [docker]
`
	os.WriteFile(path, []byte(content), 0644)

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

	content := `
- pattern: "echo hello"
  description: "Hello world"
  tags: [test]
- pattern: "ls -la"
  description: "List files"
`
	os.WriteFile(path, []byte(content), 0644)

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

func TestFindCommandFiles(t *testing.T) {
	dir := t.TempDir()

	// Create some yaml files
	os.WriteFile(filepath.Join(dir, "commands.yaml"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "extra.yml"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte(""), 0644)

	// Nested dir
	sub := filepath.Join(dir, "ops")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(sub, "runbooks.yaml"), []byte(""), 0644)

	// .git should be skipped
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

	content := `
- pattern: ""
  description: "empty"
- pattern: "real command"
  description: "valid"
`
	os.WriteFile(path, []byte(content), 0644)

	entries, err := parseCommandFile(path)
	if err != nil {
		t.Fatalf("parseCommandFile: %v", err)
	}
	// Both are returned; filtering empty patterns is the importer's job
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}
