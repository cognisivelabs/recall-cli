package main

import (
	"strings"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
)

func TestDelete_ByID_Force(t *testing.T) {
	store := newMockStore().seed(storage.Command{
		Pattern: "echo hello", Description: "test",
	})

	cmd := NewDeleteCmd(store)
	_, _, err := execCmd(cmd, "1", "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.commands) != 0 {
		t.Fatalf("expected 0 commands after delete, got %d", len(store.commands))
	}
}

func TestDelete_ByID_Confirmed(t *testing.T) {
	store := newMockStore().seed(storage.Command{
		Pattern: "echo hello", Description: "test",
	})

	cmd := NewDeleteCmd(store)
	cmd.SetIn(strings.NewReader("y\n"))

	_, stderr, err := execCmd(cmd, "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.commands) != 0 {
		t.Fatalf("expected 0 commands after delete, got %d", len(store.commands))
	}
	if !strings.Contains(stderr, "Deleted command #1") {
		t.Errorf("expected 'Deleted command #1' in stderr, got %q", stderr)
	}
}

func TestDelete_ByID_Cancelled(t *testing.T) {
	store := newMockStore().seed(storage.Command{
		Pattern: "echo hello", Description: "test",
	})

	cmd := NewDeleteCmd(store)
	cmd.SetIn(strings.NewReader("n\n"))

	_, stderr, err := execCmd(cmd, "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.commands) != 1 {
		t.Fatalf("expected 1 command (not deleted), got %d", len(store.commands))
	}
	if !strings.Contains(stderr, "Cancelled") {
		t.Errorf("expected 'Cancelled' in stderr, got %q", stderr)
	}
}

func TestDelete_ByID_NotFound(t *testing.T) {
	store := newMockStore()
	cmd := NewDeleteCmd(store)

	_, _, err := execCmd(cmd, "99")
	if err == nil {
		t.Fatal("expected error for nonexistent ID, got nil")
	}
	if !strings.Contains(err.Error(), "no command with ID 99") {
		t.Errorf("expected 'no command with ID 99', got %q", err.Error())
	}
}

func TestDelete_ByPattern_ExactMatch(t *testing.T) {
	store := newMockStore().seed(storage.Command{
		Pattern: "kubectl get pods", Description: "list pods",
	})

	cmd := NewDeleteCmd(store)
	_, _, err := execCmd(cmd, "kubectl get pods", "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.commands) != 0 {
		t.Fatalf("expected 0 commands after delete, got %d", len(store.commands))
	}
}

func TestDelete_ByPattern_NotFound(t *testing.T) {
	store := newMockStore()
	cmd := NewDeleteCmd(store)

	_, _, err := execCmd(cmd, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no command matching") {
		t.Errorf("expected 'no command matching', got %q", err.Error())
	}
}

func TestDelete_ByPattern_SubstringDoesNotMatch(t *testing.T) {
	store := newMockStore().
		seed(storage.Command{Pattern: "kubectl get pods", Description: "list pods"}).
		seed(storage.Command{Pattern: "kubectl get svc", Description: "list services"})

	cmd := NewDeleteCmd(store)
	_, _, err := execCmd(cmd, "kubectl")
	if err == nil {
		t.Fatal("expected error for non-exact match, got nil")
	}
	if !strings.Contains(err.Error(), "no command matching") {
		t.Errorf("expected 'no command matching' in error, got %q", err.Error())
	}
	if len(store.commands) != 2 {
		t.Fatalf("expected 2 commands (nothing deleted), got %d", len(store.commands))
	}
}

func TestDelete_RequiresArg(t *testing.T) {
	store := newMockStore()
	cmd := NewDeleteCmd(store)

	_, _, err := execCmd(cmd)
	if err == nil {
		t.Fatal("expected error for missing arg, got nil")
	}
}

func TestDelete_ForceSkipsPrompt(t *testing.T) {
	store := newMockStore().
		seed(storage.Command{Pattern: "echo hello", Description: "test"})

	cmd := NewDeleteCmd(store)
	_, _, err := execCmd(cmd, "echo hello", "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.commands) != 0 {
		t.Fatalf("expected 0 commands after force delete, got %d", len(store.commands))
	}
}
