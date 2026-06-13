package storage

import (
	"path/filepath"
	"testing"
)

// newTestStore creates an isolated in-memory-like store backed by a temp file.
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

// TestSQLiteStore_CRUD exercises all basic operations in sequence.
func TestSQLiteStore_CRUD(t *testing.T) {
	store := newTestStore(t)

	// List on empty store
	cmds, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(cmds) != 0 {
		t.Fatalf("expected 0 commands, got %d", len(cmds))
	}

	// Insert via Upsert
	if err := store.Upsert(Command{Pattern: "echo hello", Description: "test command", Tags: "test,dev"}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

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

	// GetByPattern — hit
	found, err := store.GetByPattern("echo hello")
	if err != nil {
		t.Fatalf("GetByPattern: %v", err)
	}
	if found == nil || found.Description != "test command" {
		t.Errorf("GetByPattern returned wrong result: %+v", found)
	}

	// GetByID — hit
	byID, err := store.GetByID(found.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if byID == nil || byID.Pattern != "echo hello" {
		t.Errorf("GetByID returned wrong result: %+v", byID)
	}

	// GetByID — miss
	notFound, err := store.GetByID(9999)
	if err != nil {
		t.Fatalf("GetByID miss: %v", err)
	}
	if notFound != nil {
		t.Fatal("expected nil for nonexistent ID")
	}

	// GetByPattern — miss
	notFoundP, err := store.GetByPattern("nonexistent")
	if err != nil {
		t.Fatalf("GetByPattern miss: %v", err)
	}
	if notFoundP != nil {
		t.Fatal("expected nil for nonexistent pattern")
	}

	// Update
	found.Description = "updated description"
	found.Tags = "updated"
	if err := store.Update(*found); err != nil {
		t.Fatalf("Update: %v", err)
	}
	updated, _ := store.GetByPattern("echo hello")
	if updated.Description != "updated description" {
		t.Errorf("Update: expected 'updated description', got %q", updated.Description)
	}

	// Upsert existing — must update, not duplicate
	if err := store.Upsert(Command{Pattern: "echo hello", Description: "upserted", Tags: "upsert"}); err != nil {
		t.Fatalf("Upsert existing: %v", err)
	}
	cmds, _ = store.List()
	if len(cmds) != 1 {
		t.Fatalf("upsert created duplicate: got %d commands", len(cmds))
	}
	if cmds[0].Description != "upserted" {
		t.Errorf("upsert did not update description, got %q", cmds[0].Description)
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
		t.Error("expected last_used_at to be set after RecordUsage")
	}

	// Delete
	if err := store.Delete(id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	cmds, _ = store.List()
	if len(cmds) != 0 {
		t.Fatalf("expected 0 after delete, got %d", len(cmds))
	}
}

// TestSQLiteStore_UpsertWithSource verifies that a non-default source is preserved.
func TestSQLiteStore_UpsertWithSource(t *testing.T) {
	store := newTestStore(t)

	store.Upsert(Command{Pattern: "kubectl get pods", Source: "team-ops"})

	cmds, _ := store.List()
	if cmds[0].Source != "team-ops" {
		t.Errorf("expected source 'team-ops', got %q", cmds[0].Source)
	}
}

// TestSQLiteStore_WorkspaceFilter verifies workspace_filter round-trips correctly.
func TestSQLiteStore_WorkspaceFilter(t *testing.T) {
	store := newTestStore(t)

	store.Upsert(Command{Pattern: "make deploy", WorkspaceFilter: "~/work/billing-*"})

	cmds, _ := store.List()
	if cmds[0].WorkspaceFilter != "~/work/billing-*" {
		t.Errorf("List: expected workspace filter, got %q", cmds[0].WorkspaceFilter)
	}

	found, _ := store.GetByPattern("make deploy")
	if found.WorkspaceFilter != "~/work/billing-*" {
		t.Errorf("GetByPattern: expected workspace_filter, got %q", found.WorkspaceFilter)
	}
}

// TestSQLiteStore_ListOrderByUsage verifies that List returns most-used commands first.
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

// TestSQLiteStore_TagsRoundTrip verifies that tags stored via Upsert and Update
// are returned verbatim — the single comma-separated column is the source of truth.
func TestSQLiteStore_TagsRoundTrip(t *testing.T) {
	store := newTestStore(t)

	store.Upsert(Command{Pattern: "docker ps", Tags: "docker,containers,dev"})
	found, _ := store.GetByPattern("docker ps")
	if found.Tags != "docker,containers,dev" {
		t.Errorf("Upsert: tags round-trip failed, got %q", found.Tags)
	}

	found.Tags = "docker,prod"
	store.Update(*found)
	updated, _ := store.GetByPattern("docker ps")
	if updated.Tags != "docker,prod" {
		t.Errorf("Update: tags round-trip failed, got %q", updated.Tags)
	}
}

// TestSQLiteStore_UpsertPreservesUsageCount verifies that upserting an existing
// command does not reset usage_count or last_used_at.
func TestSQLiteStore_UpsertPreservesUsageCount(t *testing.T) {
	store := newTestStore(t)

	store.Upsert(Command{Pattern: "echo hi"})
	cmds, _ := store.List()
	store.RecordUsage(cmds[0].ID)
	store.RecordUsage(cmds[0].ID)

	// Upsert again (same pattern, different description)
	store.Upsert(Command{Pattern: "echo hi", Description: "new desc"})

	cmds, _ = store.List()
	if cmds[0].UsageCount != 2 {
		t.Errorf("upsert must not reset usage_count, got %d", cmds[0].UsageCount)
	}
	if cmds[0].LastUsedAt.IsZero() {
		t.Error("upsert must not reset last_used_at")
	}
}

// TestSQLiteStore_DeleteNonexistent verifies Delete is a no-op for unknown IDs.
func TestSQLiteStore_DeleteNonexistent(t *testing.T) {
	store := newTestStore(t)
	if err := store.Delete(9999); err != nil {
		t.Errorf("Delete of nonexistent ID should not error, got: %v", err)
	}
}

// TestSQLiteStore_MultipleCommands verifies List returns all rows and ordering.
func TestSQLiteStore_MultipleCommands(t *testing.T) {
	store := newTestStore(t)

	for _, p := range []string{"echo a", "echo b", "echo c"} {
		store.Upsert(Command{Pattern: p})
	}

	cmds, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(cmds) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(cmds))
	}
}

// TestSQLiteStore_SchemaIdempotent verifies that opening the same file twice
// does not error (IF NOT EXISTS guards must be in place).
func TestSQLiteStore_SchemaIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	s1, err := NewSQLiteStoreAt(dbPath)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	s1.Close()

	s2, err := NewSQLiteStoreAt(dbPath)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	s2.Close()
}
