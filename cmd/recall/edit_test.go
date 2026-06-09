package main

import (
	"strings"
	"testing"
)

func TestEdit_ByID_NotFound(t *testing.T) {
	store := newMockStore()
	cmd := NewEditCmd(store)

	_, _, err := execCmd(cmd, "99")
	if err == nil {
		t.Fatal("expected error for nonexistent ID, got nil")
	}
	if !strings.Contains(err.Error(), "no command with ID 99") {
		t.Errorf("expected 'no command with ID 99', got %q", err.Error())
	}
}

func TestEdit_ByPattern_NotFound(t *testing.T) {
	store := newMockStore()
	cmd := NewEditCmd(store)

	_, _, err := execCmd(cmd, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent pattern, got nil")
	}
	if !strings.Contains(err.Error(), "no command matching") {
		t.Errorf("expected 'no command matching', got %q", err.Error())
	}
}

func TestEdit_RequiresArg(t *testing.T) {
	store := newMockStore()
	cmd := NewEditCmd(store)

	_, _, err := execCmd(cmd)
	if err == nil {
		t.Fatal("expected error for missing arg, got nil")
	}
}
