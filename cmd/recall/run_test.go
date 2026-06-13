package main

import (
	"strings"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
	"github.com/spf13/cobra"
)

// sentinel returns a cobra.Command with default I/O buffers for test use.
func sentinelCmd() *cobra.Command {
	cmd := &cobra.Command{}
	var sb strings.Builder
	cmd.SetOut(&sb)
	cmd.SetErr(&sb)
	return cmd
}

// --- exitCodeError ---

func TestExitCodeError_ErrorString(t *testing.T) {
	err := exitCodeError(1)
	if err.Error() != "exit status 1" {
		t.Errorf("unexpected error string %q", err.Error())
	}
	err2 := exitCodeError(127)
	if err2.Error() != "exit status 127" {
		t.Errorf("unexpected error string %q", err2.Error())
	}
}

// --- findBestMatch ---

func TestFindBestMatch_ReturnsTopResult(t *testing.T) {
	store := newMockStore().
		seed(storage.Command{Pattern: "kubectl get pods", Tags: "k8s"}).
		seed(storage.Command{Pattern: "docker ps"})

	match, err := findBestMatch(store, "kubectl", "", sentinelCmd())
	if err != nil {
		t.Fatalf("findBestMatch: %v", err)
	}
	if match == nil {
		t.Fatal("expected a match, got nil")
	}
	if match.Pattern != "kubectl get pods" {
		t.Errorf("expected 'kubectl get pods', got %q", match.Pattern)
	}
}

func TestFindBestMatch_NoResults(t *testing.T) {
	store := newMockStore().
		seed(storage.Command{Pattern: "docker ps"})

	match, err := findBestMatch(store, "xyzzy_no_match", "", sentinelCmd())
	if err != nil {
		t.Fatalf("findBestMatch: %v", err)
	}
	if match != nil {
		t.Errorf("expected nil match, got %q", match.Pattern)
	}
}

func TestFindBestMatch_FiltersByTag(t *testing.T) {
	store := newMockStore().
		seed(storage.Command{Pattern: "kubectl get pods", Tags: "k8s"}).
		seed(storage.Command{Pattern: "kubectl rollout", Tags: "deploy"})

	match, err := findBestMatch(store, "kubectl", "k8s", sentinelCmd())
	if err != nil {
		t.Fatalf("findBestMatch: %v", err)
	}
	if match == nil {
		t.Fatal("expected a match, got nil")
	}
	if match.Pattern != "kubectl get pods" {
		t.Errorf("expected k8s-tagged command, got %q", match.Pattern)
	}
}

func TestFindBestMatch_TieNotice(t *testing.T) {
	// Seed two commands with identical patterns to guarantee a scoring tie.
	store := newMockStore().
		seed(storage.Command{Pattern: "echo aaa", Description: "same score"}).
		seed(storage.Command{Pattern: "echo bbb", Description: "same score"})

	var errBuf strings.Builder
	cmd := &cobra.Command{}
	cmd.SetErr(&errBuf)
	cmd.SetOut(&strings.Builder{})

	match, err := findBestMatch(store, "echo", "", cmd)
	if err != nil {
		t.Fatalf("findBestMatch: %v", err)
	}
	if match == nil {
		t.Fatal("expected a match, got nil")
	}
	// When the top two results tie, an ambiguity notice is printed to stderr.
	if !strings.Contains(errBuf.String(), "echo") {
		t.Errorf("expected ambiguity notice on stderr mentioning the matched commands, got %q", errBuf.String())
	}
}

// --- resolvePlaceholders (no-TUI paths) ---

func TestResolvePlaceholders_NoPlaceholders(t *testing.T) {
	cmd := sentinelCmd()
	resolved, cancelled, err := resolvePlaceholders(cmd, "echo hello world")
	if err != nil {
		t.Fatalf("resolvePlaceholders: %v", err)
	}
	if cancelled {
		t.Error("expected not cancelled")
	}
	if resolved != "echo hello world" {
		t.Errorf("expected unchanged pattern, got %q", resolved)
	}
}

func TestResolvePlaceholders_AutoResolvedOnly(t *testing.T) {
	// {{cwd}} is an auto-resolved placeholder — no TUI required.
	cmd := sentinelCmd()
	resolved, cancelled, err := resolvePlaceholders(cmd, "ls {{cwd}}")
	if err != nil {
		t.Fatalf("resolvePlaceholders: %v", err)
	}
	if cancelled {
		t.Error("expected not cancelled")
	}
	// {{cwd}} should be replaced with a real path; no literal braces should remain.
	if strings.Contains(resolved, "{{cwd}}") {
		t.Errorf("expected {{cwd}} to be resolved, got %q", resolved)
	}
}
