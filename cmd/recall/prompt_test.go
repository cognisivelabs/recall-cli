package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
)

func TestConfirmFromReader_Yes(t *testing.T) {
	for _, input := range []string{"y\n", "yes\n", "Y\n", "YES\n", "  y  \n"} {
		result := confirmFromReader(strings.NewReader(input), &bytes.Buffer{}, "confirm? ")
		if !result {
			t.Errorf("expected true for input %q", input)
		}
	}
}

func TestConfirmFromReader_No(t *testing.T) {
	for _, input := range []string{"n\n", "no\n", "\n", "x\n", ""} {
		result := confirmFromReader(strings.NewReader(input), &bytes.Buffer{}, "confirm? ")
		if result {
			t.Errorf("expected false for input %q", input)
		}
	}
}

func TestConfirmFromReader_WritesPrompt(t *testing.T) {
	out := &bytes.Buffer{}
	confirmFromReader(strings.NewReader("n\n"), out, "Delete? [y/N] ")
	if out.String() != "Delete? [y/N] " {
		t.Errorf("expected prompt written to output, got %q", out.String())
	}
}

func TestPrintMatchList(t *testing.T) {
	out := &bytes.Buffer{}
	matches := []storage.Command{
		{ID: 1, Pattern: "echo hello", Description: "test"},
		{ID: 2, Pattern: "echo world", Description: "another"},
	}
	printMatchList(out, "echo", matches)

	output := out.String()
	if !strings.Contains(output, `Multiple matches for "echo"`) {
		t.Errorf("expected header, got %q", output)
	}
	if !strings.Contains(output, "echo hello") || !strings.Contains(output, "echo world") {
		t.Errorf("expected both matches listed, got %q", output)
	}
}

func TestFindCommandByIDOrPattern_ByID(t *testing.T) {
	store := newMockStore().seed(storage.Command{
		Pattern: "echo hello", Description: "test",
	})

	result, err := findCommandByIDOrPattern(store, "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Pattern != "echo hello" {
		t.Errorf("expected 'echo hello', got %q", result.Pattern)
	}
}

func TestFindCommandByIDOrPattern_ByPattern(t *testing.T) {
	store := newMockStore().seed(storage.Command{
		Pattern: "kubectl get pods", Description: "list pods",
	})

	result, err := findCommandByIDOrPattern(store, "kubectl get pods")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Description != "list pods" {
		t.Errorf("expected 'list pods', got %q", result.Description)
	}
}

func TestFindCommandByIDOrPattern_IDNotFound(t *testing.T) {
	store := newMockStore()

	_, err := findCommandByIDOrPattern(store, "99")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no command with ID 99") {
		t.Errorf("expected 'no command with ID 99', got %q", err.Error())
	}
}

func TestFindCommandByIDOrPattern_PatternNotFound(t *testing.T) {
	store := newMockStore()

	_, err := findCommandByIDOrPattern(store, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no command matching") {
		t.Errorf("expected 'no command matching', got %q", err.Error())
	}
}
