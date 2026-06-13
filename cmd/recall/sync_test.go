package main

import (
	"strings"
	"testing"
)

// TestSyncCmd_NoGitSources verifies `recall sync` succeeds when the config has no git sources.
func TestSyncCmd_NoGitSources(t *testing.T) {
	isolateCmdEnv(t)

	store := newMockStore()
	cmd := NewSyncCmd(store)
	_, _, err := execCmd(cmd)
	if err != nil {
		t.Fatalf("sync with no git sources: %v", err)
	}
}

// TestSyncCmd_PrintsCompletionMessage verifies "Sync complete." is written to stderr.
func TestSyncCmd_PrintsCompletionMessage(t *testing.T) {
	isolateCmdEnv(t)

	store := newMockStore()
	cmd := NewSyncCmd(store)
	_, stderr, err := execCmd(cmd)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if !strings.Contains(stderr, "Sync complete.") {
		t.Errorf("expected 'Sync complete.' in stderr, got %q", stderr)
	}
}
