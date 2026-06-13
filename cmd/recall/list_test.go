package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
)

// --- truncate ---

func TestTruncate_ShortString(t *testing.T) {
	got := truncate("hello", 10)
	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestTruncate_ExactLength(t *testing.T) {
	got := truncate("hello", 5)
	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestTruncate_LongString(t *testing.T) {
	got := truncate("hello world foo bar", 10)
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected ellipsis, got %q", got)
	}
	if len([]rune(got)) != 10 {
		t.Errorf("expected length 10, got %d", len([]rune(got)))
	}
}

func TestTruncate_Unicode(t *testing.T) {
	// Each Japanese character is one rune, not one byte.
	s := "日本語テスト長い文字列"
	got := truncate(s, 6)
	if len([]rune(got)) != 6 {
		t.Errorf("expected 6 runes, got %d", len([]rune(got)))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected ellipsis, got %q", got)
	}
}

// --- printTable ---

func TestPrintTable_Headers(t *testing.T) {
	var b strings.Builder
	printTable(&b, []storage.Command{
		{ID: 1, Pattern: "echo hi", Description: "test", Tags: "dev", Source: "local"},
	})
	out := b.String()
	for _, header := range []string{"ID", "COMMAND", "DESCRIPTION", "TAGS", "USED", "SOURCE"} {
		if !strings.Contains(out, header) {
			t.Errorf("table missing header %q", header)
		}
	}
}

func TestPrintTable_ContainsValues(t *testing.T) {
	var b strings.Builder
	printTable(&b, []storage.Command{
		{ID: 42, Pattern: "kubectl get pods", Description: "list pods", Tags: "k8s", Source: "local", UsageCount: 7},
	})
	out := b.String()
	for _, want := range []string{"42", "kubectl get pods", "list pods", "k8s", "local", "7"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing value %q\nfull output:\n%s", want, out)
		}
	}
}

func TestPrintTable_TruncatesLongPattern(t *testing.T) {
	var b strings.Builder
	long := strings.Repeat("a", 60)
	printTable(&b, []storage.Command{
		{ID: 1, Pattern: long},
	})
	if strings.Contains(b.String(), long) {
		t.Error("expected long pattern to be truncated")
	}
	if !strings.Contains(b.String(), "...") {
		t.Error("expected ellipsis for truncated pattern")
	}
}

// --- printJSON ---

func TestPrintJSON_ValidJSON(t *testing.T) {
	var b strings.Builder
	cmds := []storage.Command{
		{ID: 1, Pattern: "echo hi", Description: "test", Tags: "dev", Source: "local", UsageCount: 3},
	}
	if err := printJSON(&b, cmds); err != nil {
		t.Fatalf("printJSON error: %v", err)
	}

	var parsed []commandJSON
	if err := json.Unmarshal([]byte(b.String()), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, b.String())
	}
	if len(parsed) != 1 {
		t.Fatalf("expected 1 element, got %d", len(parsed))
	}
	if parsed[0].Pattern != "echo hi" {
		t.Errorf("expected pattern 'echo hi', got %q", parsed[0].Pattern)
	}
	if parsed[0].UsageCount != 3 {
		t.Errorf("expected usage_count 3, got %d", parsed[0].UsageCount)
	}
}

func TestPrintJSON_SpecialCharsEscaped(t *testing.T) {
	var b strings.Builder
	cmds := []storage.Command{
		{ID: 1, Pattern: `echo "hello\nworld"`},
	}
	if err := printJSON(&b, cmds); err != nil {
		t.Fatalf("printJSON error: %v", err)
	}
	// Must parse without error — encoding/json handles escaping correctly.
	var parsed []commandJSON
	if err := json.Unmarshal([]byte(b.String()), &parsed); err != nil {
		t.Fatalf("special chars broke JSON: %v\noutput: %s", err, b.String())
	}
}

func TestPrintJSON_EmptySlice(t *testing.T) {
	var b strings.Builder
	if err := printJSON(&b, []storage.Command{}); err != nil {
		t.Fatalf("printJSON error: %v", err)
	}
	var parsed []commandJSON
	if err := json.Unmarshal([]byte(b.String()), &parsed); err != nil {
		t.Fatalf("empty slice is not valid JSON: %v\noutput: %s", err, b.String())
	}
	if len(parsed) != 0 {
		t.Errorf("expected 0 elements, got %d", len(parsed))
	}
}

// --- NewListCmd integration ---

func TestListCmd_TableOutput(t *testing.T) {
	store := newMockStore().
		seed(storage.Command{Pattern: "kubectl get pods", Description: "list pods", Tags: "k8s", Source: "local"})

	cmd := NewListCmd(store)
	stdout, _, err := execCmd(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "kubectl get pods") {
		t.Errorf("expected pattern in table output, got:\n%s", stdout)
	}
}

func TestListCmd_JSONOutput(t *testing.T) {
	store := newMockStore().
		seed(storage.Command{Pattern: "docker ps", Description: "containers", Tags: "docker"})

	cmd := NewListCmd(store)
	stdout, _, err := execCmd(cmd, "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []commandJSON
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("--json output not valid JSON: %v\noutput: %s", err, stdout)
	}
	if len(parsed) != 1 || parsed[0].Pattern != "docker ps" {
		t.Errorf("unexpected JSON content: %+v", parsed)
	}
}

func TestListCmd_FilterByTag(t *testing.T) {
	store := newMockStore().
		seed(storage.Command{Pattern: "kubectl get pods", Tags: "k8s"}).
		seed(storage.Command{Pattern: "docker ps", Tags: "docker"})

	cmd := NewListCmd(store)
	stdout, _, err := execCmd(cmd, "-t", "k8s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "kubectl get pods") {
		t.Errorf("expected k8s command in output")
	}
	if strings.Contains(stdout, "docker ps") {
		t.Errorf("docker command should be filtered out")
	}
}

func TestListCmd_FilterBySource(t *testing.T) {
	store := newMockStore().
		seed(storage.Command{Pattern: "kubectl get pods", Source: "team-ops"}).
		seed(storage.Command{Pattern: "docker ps", Source: "local"})

	cmd := NewListCmd(store)
	stdout, _, err := execCmd(cmd, "-s", "team-ops")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "kubectl get pods") {
		t.Errorf("expected team-ops command in output")
	}
	if strings.Contains(stdout, "docker ps") {
		t.Errorf("local command should be filtered out")
	}
}

func TestListCmd_EmptyStore(t *testing.T) {
	store := newMockStore()
	cmd := NewListCmd(store)
	_, stderr, err := execCmd(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "No commands found") {
		t.Errorf("expected 'No commands found' in stderr, got %q", stderr)
	}
}

func TestListCmd_FilterReturnsNothing(t *testing.T) {
	store := newMockStore().
		seed(storage.Command{Pattern: "docker ps", Tags: "docker"})

	cmd := NewListCmd(store)
	_, stderr, err := execCmd(cmd, "-t", "k8s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "No commands found") {
		t.Errorf("expected 'No commands found' for unmatched filter, got %q", stderr)
	}
}
