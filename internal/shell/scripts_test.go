package shell

import (
	"strings"
	"testing"
)

// TestPrintInitScript_Zsh verifies the zsh script is non-empty and contains
// the widget function name that the shell integration relies on.
func TestPrintInitScript_Zsh(t *testing.T) {
	var b strings.Builder
	if err := PrintInitScript(&b, "zsh"); err != nil {
		t.Fatalf("PrintInitScript zsh: %v", err)
	}
	out := b.String()
	if !strings.Contains(out, "recall_widget") {
		t.Error("zsh script missing 'recall_widget' function")
	}
	if len(out) == 0 {
		t.Error("zsh script is empty")
	}
}

// TestPrintInitScript_Bash verifies the bash script is non-empty and contains
// a reference to recall.
func TestPrintInitScript_Bash(t *testing.T) {
	var b strings.Builder
	if err := PrintInitScript(&b, "bash"); err != nil {
		t.Fatalf("PrintInitScript bash: %v", err)
	}
	out := b.String()
	if len(out) == 0 {
		t.Error("bash script is empty")
	}
	if !strings.Contains(out, "recall") {
		t.Error("bash script does not mention 'recall'")
	}
}

// TestPrintInitScript_UnsupportedShell verifies that an unknown shell returns
// an error instead of silently emitting a fallback script.
func TestPrintInitScript_UnsupportedShell(t *testing.T) {
	var b strings.Builder
	err := PrintInitScript(&b, "fish")
	if err == nil {
		t.Fatal("expected error for unsupported shell 'fish', got nil")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected 'unsupported' in error message, got %q", err.Error())
	}
	if b.Len() != 0 {
		t.Error("no output should be written for an unsupported shell")
	}
}

// TestPrintInitScript_EmptyShell verifies that an empty shell string returns an error.
func TestPrintInitScript_EmptyShell(t *testing.T) {
	var b strings.Builder
	if err := PrintInitScript(&b, ""); err == nil {
		t.Fatal("expected error for empty shell string, got nil")
	}
}
