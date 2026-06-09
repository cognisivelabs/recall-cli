package main

import (
	"strings"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
)

func TestAdd_NonInteractive_NewCommand(t *testing.T) {
	store := newMockStore()
	cmd := NewAddCmd(store)

	stdout, _, err := execCmd(cmd, "echo hello", "-d", "test command", "-t", "dev,test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(store.commands))
	}
	if store.commands[0].Pattern != "echo hello" {
		t.Errorf("expected pattern 'echo hello', got %q", store.commands[0].Pattern)
	}
	if store.commands[0].Description != "test command" {
		t.Errorf("expected description 'test command', got %q", store.commands[0].Description)
	}
	if store.commands[0].Tags != "dev,test" {
		t.Errorf("expected tags 'dev,test', got %q", store.commands[0].Tags)
	}
	if !strings.Contains(stdout, "Saved") {
		t.Errorf("expected 'Saved' in stdout, got %q", stdout)
	}
}

func TestAdd_NonInteractive_UpdateExisting(t *testing.T) {
	store := newMockStore().seed(storage.Command{
		Pattern: "echo hello", Description: "old desc", Tags: "old",
	})

	cmd := NewAddCmd(store)
	stdout, _, err := execCmd(cmd, "echo hello", "-d", "new desc", "-t", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.commands) != 1 {
		t.Fatalf("expected 1 command (upsert, not duplicate), got %d", len(store.commands))
	}
	if store.commands[0].Description != "new desc" {
		t.Errorf("expected description 'new desc', got %q", store.commands[0].Description)
	}
	if !strings.Contains(stdout, "Updated") {
		t.Errorf("expected 'Updated' in stdout, got %q", stdout)
	}
}

func TestAdd_NonInteractive_TagsOnly(t *testing.T) {
	store := newMockStore()
	cmd := NewAddCmd(store)

	_, _, err := execCmd(cmd, "kubectl get pods", "-t", "k8s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(store.commands))
	}
	if store.commands[0].Tags != "k8s" {
		t.Errorf("expected tags 'k8s', got %q", store.commands[0].Tags)
	}
}

func TestAdd_NonInteractive_ClearsTags(t *testing.T) {
	store := newMockStore().seed(storage.Command{
		Pattern: "echo hello", Description: "test", Tags: "old-tag",
	})

	cmd := NewAddCmd(store)
	_, _, err := execCmd(cmd, "echo hello", "-d", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.commands[0].Tags != "" {
		t.Errorf("expected tags cleared, got %q", store.commands[0].Tags)
	}
}

func TestAdd_NonInteractive_TrimsWhitespace(t *testing.T) {
	store := newMockStore()
	cmd := NewAddCmd(store)

	_, _, err := execCmd(cmd, "  echo hello  ", "-d", "  some desc  ", "-t", "  a, b  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.commands[0].Pattern != "echo hello" {
		t.Errorf("expected trimmed pattern, got %q", store.commands[0].Pattern)
	}
	if store.commands[0].Description != "some desc" {
		t.Errorf("expected trimmed description, got %q", store.commands[0].Description)
	}
	if store.commands[0].Tags != "a, b" {
		t.Errorf("expected trimmed tags, got %q", store.commands[0].Tags)
	}
}

func TestAdd_Interactive_RejectsDuplicate(t *testing.T) {
	store := newMockStore().seed(storage.Command{
		Pattern: "echo hello", Description: "existing",
	})

	cmd := NewAddCmd(store)
	_, _, err := execCmd(cmd, "echo hello")
	if err == nil {
		t.Fatal("expected error for duplicate, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "recall edit") {
		t.Errorf("expected 'recall edit' hint in error, got %q", err.Error())
	}
}
