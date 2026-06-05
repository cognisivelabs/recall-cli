package storage

import (
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewSQLiteStoreAt(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStoreAt: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestSQLiteStore_CRUD(t *testing.T) {
	store := newTestStore(t)

	// List should be empty
	cmds, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(cmds) != 0 {
		t.Fatalf("expected 0 commands, got %d", len(cmds))
	}

	// Upsert (insert)
	err = store.Upsert(Command{
		Pattern:     "echo hello",
		Description: "test command",
		Tags:        "test,dev",
	})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	// List should have 1
	cmds, err = store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Pattern != "echo hello" {
		t.Errorf("expected pattern 'echo hello', got %q", cmds[0].Pattern)
	}
	if cmds[0].Source != "local" {
		t.Errorf("expected source 'local', got %q", cmds[0].Source)
	}

	// GetByPattern
	found, err := store.GetByPattern("echo hello")
	if err != nil {
		t.Fatalf("GetByPattern: %v", err)
	}
	if found == nil {
		t.Fatal("expected to find command")
	}
	if found.Description != "test command" {
		t.Errorf("expected description 'test command', got %q", found.Description)
	}

	// GetByID
	byID, err := store.GetByID(found.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if byID == nil {
		t.Fatal("expected to find command by ID")
	}
	if byID.Pattern != "echo hello" {
		t.Errorf("expected pattern 'echo hello', got %q", byID.Pattern)
	}

	// GetByID not found
	notFound, err := store.GetByID(9999)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if notFound != nil {
		t.Fatal("expected nil for nonexistent ID")
	}

	// GetByPattern not found
	notFoundP, err := store.GetByPattern("nonexistent")
	if err != nil {
		t.Fatalf("GetByPattern: %v", err)
	}
	if notFoundP != nil {
		t.Fatal("expected nil for nonexistent pattern")
	}

	// Update
	found.Description = "updated description"
	found.Tags = "updated"
	err = store.Update(*found)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	updated, _ := store.GetByPattern("echo hello")
	if updated.Description != "updated description" {
		t.Errorf("expected updated description, got %q", updated.Description)
	}

	// Upsert existing (should update, not duplicate)
	err = store.Upsert(Command{
		Pattern:     "echo hello",
		Description: "upserted",
		Tags:        "upsert",
	})
	if err != nil {
		t.Fatalf("Upsert existing: %v", err)
	}
	cmds, _ = store.List()
	if len(cmds) != 1 {
		t.Fatalf("upsert created duplicate: got %d commands", len(cmds))
	}
	if cmds[0].Description != "upserted" {
		t.Errorf("expected upserted description, got %q", cmds[0].Description)
	}

	// RecordUsage
	id := cmds[0].ID
	store.RecordUsage(id)
	store.RecordUsage(id)
	cmds, _ = store.List()
	if cmds[0].UsageCount != 2 {
		t.Errorf("expected usage_count 2, got %d", cmds[0].UsageCount)
	}
	if cmds[0].LastUsedAt.IsZero() {
		t.Error("expected last_used_at to be set")
	}

	// Delete
	err = store.Delete(id)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	cmds, _ = store.List()
	if len(cmds) != 0 {
		t.Fatalf("expected 0 after delete, got %d", len(cmds))
	}
}

func TestSQLiteStore_UpsertWithSource(t *testing.T) {
	store := newTestStore(t)

	store.Upsert(Command{
		Pattern: "kubectl get pods",
		Source:  "team-ops",
	})

	cmds, _ := store.List()
	if cmds[0].Source != "team-ops" {
		t.Errorf("expected source 'team-ops', got %q", cmds[0].Source)
	}
}

func TestSQLiteStore_WorkspaceFilter(t *testing.T) {
	store := newTestStore(t)

	store.Upsert(Command{
		Pattern:         "make deploy",
		WorkspaceFilter: "~/work/billing-*",
	})

	cmds, _ := store.List()
	if cmds[0].WorkspaceFilter != "~/work/billing-*" {
		t.Errorf("expected workspace filter, got %q", cmds[0].WorkspaceFilter)
	}

	// GetByPattern should also return workspace
	found, _ := store.GetByPattern("make deploy")
	if found.WorkspaceFilter != "~/work/billing-*" {
		t.Errorf("GetByPattern missing workspace_filter, got %q", found.WorkspaceFilter)
	}
}

func TestSQLiteStore_ListOrderByUsage(t *testing.T) {
	store := newTestStore(t)

	store.Upsert(Command{Pattern: "cmd-a"})
	store.Upsert(Command{Pattern: "cmd-b"})

	cmds, _ := store.List()
	var bID int
	for _, c := range cmds {
		if c.Pattern == "cmd-b" {
			bID = c.ID
		}
	}
	store.RecordUsage(bID)
	store.RecordUsage(bID)
	store.RecordUsage(bID)

	cmds, _ = store.List()
	if cmds[0].Pattern != "cmd-b" {
		t.Errorf("expected cmd-b first (most used), got %q", cmds[0].Pattern)
	}
}
