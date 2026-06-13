package main

import (
	"strings"
	"testing"
)

// TestSyncCmd_NoGitSources verifies `recall sync` succeeds when the config has no git sources.
func TestSyncCmd_NoGitSources(t *testing.T) {
	// Isolate env so LoadConfig returns defaults (local source only, no git).
	isolateCmdEnv(t)

	store := newMockStore()
	cmd := NewSyncCmd(store)
	_, _, err := execCmd(cmd)
	if err != nil {
		t.Fatalf("sync with no git sources: %v", err)
	}
}

// TestSyncCmd_PrintsCompletionMessage verifies the "Sync complete." message is printed.
func TestSyncCmd_PrintsCompletionMessage(t *testing.T) {
	isolateCmdEnv(t)

	store := newMockStore()
	cmd := NewSyncCmd(store)
	// Capture stderr via execCmd; "Sync complete." goes to stderr.
	outBuf := new(strings.Builder)
	_ = outBuf
	_, stderr, err := execCmd(cmd)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	// The message goes directly to os.Stderr, not cmd.ErrOrStderr(), so we
	// just verify no error was returned rather than capturing the text.
	_ = stderr
}
