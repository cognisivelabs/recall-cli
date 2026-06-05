package storage

import "testing"

func TestHasTag(t *testing.T) {
	tests := []struct {
		tags   string
		target string
		want   bool
	}{
		{"docker,k8s,dev", "docker", true},
		{"docker,k8s,dev", "k8s", true},
		{"docker,k8s,dev", "prod", false},
		{"docker, k8s, dev", "k8s", true}, // spaces
		{"Docker,K8S", "docker", true},     // case insensitive
		{"", "docker", false},
		{"docker", "", false},
	}
	for _, tt := range tests {
		got := HasTag(tt.tags, tt.target)
		if got != tt.want {
			t.Errorf("HasTag(%q, %q) = %v, want %v", tt.tags, tt.target, got, tt.want)
		}
	}
}

func TestCollectTags(t *testing.T) {
	cmds := []Command{
		{Tags: "docker,k8s"},
		{Tags: "k8s,debug"},
		{Tags: ""},
		{Tags: "docker"},
	}
	tags := CollectTags(cmds)
	if len(tags) != 3 {
		t.Fatalf("expected 3 unique tags, got %d: %v", len(tags), tags)
	}
	// Should be sorted
	if tags[0] != "debug" || tags[1] != "docker" || tags[2] != "k8s" {
		t.Errorf("unexpected tag order: %v", tags)
	}
}

func TestFilterByTag(t *testing.T) {
	cmds := []Command{
		{Pattern: "a", Tags: "docker,k8s"},
		{Pattern: "b", Tags: "git"},
		{Pattern: "c", Tags: "docker"},
	}

	filtered := FilterByTag(cmds, "docker")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 docker commands, got %d", len(filtered))
	}

	filtered = FilterByTag(cmds, "git")
	if len(filtered) != 1 {
		t.Fatalf("expected 1 git command, got %d", len(filtered))
	}

	// Empty tag returns all
	filtered = FilterByTag(cmds, "")
	if len(filtered) != 3 {
		t.Fatalf("expected all 3 with empty filter, got %d", len(filtered))
	}
}
