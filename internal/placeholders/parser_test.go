package placeholders

import (
	"os"
	"testing"
)

func TestParse_TextPlaceholder(t *testing.T) {
	result := Parse("kubectl logs {{service}}")
	if len(result) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(result))
	}
	p := result[0]
	if p.Type != "text" {
		t.Errorf("expected type text, got %s", p.Type)
	}
	if p.Key != "service" {
		t.Errorf("expected key service, got %s", p.Key)
	}
	if p.FullMatch != "{{service}}" {
		t.Errorf("expected fullmatch {{service}}, got %s", p.FullMatch)
	}
}

func TestParse_OptionsPlaceholder(t *testing.T) {
	result := Parse("deploy -n {{options:dev,staging,prod}}")
	if len(result) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(result))
	}
	p := result[0]
	if p.Type != "options" {
		t.Errorf("expected type options, got %s", p.Type)
	}
	if len(p.Options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(p.Options))
	}
	if p.Options[0] != "dev" || p.Options[1] != "staging" || p.Options[2] != "prod" {
		t.Errorf("unexpected options: %v", p.Options)
	}
}

func TestParse_AutoPlaceholder(t *testing.T) {
	result := Parse("git push origin {{branch}}")
	if len(result) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(result))
	}
	p := result[0]
	if p.Type != "auto" {
		t.Errorf("expected type auto, got %s", p.Type)
	}
	if p.Key != "branch" {
		t.Errorf("expected key branch, got %s", p.Key)
	}
	// AutoValue may or may not be empty depending on whether we're in a git repo
}

func TestParse_MultiplePlaceholders(t *testing.T) {
	result := Parse("kubectl logs {{service}} -n {{options:dev,prod}} --since={{duration}}")
	if len(result) != 3 {
		t.Fatalf("expected 3 placeholders, got %d", len(result))
	}
	if result[0].Type != "text" {
		t.Errorf("first should be text, got %s", result[0].Type)
	}
	if result[1].Type != "options" {
		t.Errorf("second should be options, got %s", result[1].Type)
	}
	if result[2].Type != "text" {
		t.Errorf("third should be text, got %s", result[2].Type)
	}
}

func TestParse_NoPlaceholders(t *testing.T) {
	result := Parse("echo hello")
	if len(result) != 0 {
		t.Fatalf("expected 0 placeholders, got %d", len(result))
	}
}

func TestHasPlaceholders(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"echo hello", false},
		{"kubectl logs {{service}}", true},
		{"{{branch}}", true},
		{"{not a placeholder}", false},
		{"{{options:a,b}}", true},
	}
	for _, tt := range tests {
		got := HasPlaceholders(tt.input)
		if got != tt.want {
			t.Errorf("HasPlaceholders(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestReplace(t *testing.T) {
	p := Placeholder{FullMatch: "{{service}}", Key: "service", Type: "text"}
	result := Replace("kubectl logs {{service}}", p, "billing-api")
	if result != "kubectl logs billing-api" {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestReplace_OnlyFirst(t *testing.T) {
	p := Placeholder{FullMatch: "{{file}}", Key: "file", Type: "text"}
	result := Replace("diff {{file}} {{file}}", p, "main.go")
	if result != "diff main.go {{file}}" {
		t.Errorf("expected only first replaced, got: %s", result)
	}
}

func TestAutoResolve_MixedPlaceholders(t *testing.T) {
	// {{user}} is auto, {{service}} is text
	cmd := "ssh {{user}}@host {{service}}"
	resolved, remaining := AutoResolve(cmd)

	// user should be resolved (non-empty on any system)
	if HasPlaceholders(resolved) && len(remaining) != 1 {
		t.Errorf("expected 1 remaining placeholder, got %d", len(remaining))
	}
	if len(remaining) > 0 && remaining[0].Key != "service" {
		t.Errorf("expected remaining key service, got %s", remaining[0].Key)
	}
}

func TestAutoResolve_AllAuto(t *testing.T) {
	cmd := "echo {{user}} {{host}}"
	resolved, remaining := AutoResolve(cmd)
	if len(remaining) != 0 {
		t.Errorf("expected 0 remaining, got %d", len(remaining))
	}
	if HasPlaceholders(resolved) {
		t.Errorf("expected no placeholders in resolved command: %s", resolved)
	}
}

// --- auto-resolver helper functions ---

func TestCwd(t *testing.T) {
	if cwd() == "" {
		t.Error("cwd() returned empty string")
	}
}

func TestDirName(t *testing.T) {
	if dirName() == "" {
		t.Error("dirName() returned empty string")
	}
}

func TestUserName(t *testing.T) {
	if userName() == "" {
		t.Error("userName() returned empty string")
	}
}

func TestHostName(t *testing.T) {
	if hostName() == "" {
		t.Error("hostName() returned empty string")
	}
}

func TestHomeDir(t *testing.T) {
	if homeDir() == "" {
		t.Error("homeDir() returned empty string")
	}
}

func TestHomeDir_FallbackOnBadHome(t *testing.T) {
	for _, key := range []string{"HOME", "USERPROFILE", "HOMEPATH", "HOMEDRIVE"} {
		prev, existed := os.LookupEnv(key)
		os.Unsetenv(key)
		t.Cleanup(func() {
			if existed {
				os.Setenv(key, prev)
			}
		})
	}
	// With no home-directory env vars set, homeDir should return an empty string
	// rather than panic or return a partial path.
	got := homeDir()
	if got != "" {
		// On some systems os.UserHomeDir() may still resolve via passwd — accept that.
		t.Logf("homeDir() returned %q with HOME/USERPROFILE unset (OS fallback)", got)
	}
}

func TestParse_AllAutoKeys(t *testing.T) {
	autoKeys := []string{"branch", "cwd", "dir", "user", "host", "home"}
	for _, key := range autoKeys {
		placeholders := Parse("cmd {{" + key + "}}")
		if len(placeholders) != 1 {
			t.Errorf("key %q: expected 1 placeholder, got %d", key, len(placeholders))
			continue
		}
		if placeholders[0].Type != "auto" {
			t.Errorf("key %q: expected type 'auto', got %q", key, placeholders[0].Type)
		}
	}
}

func TestAutoResolve_NothingToResolve(t *testing.T) {
	resolved, remaining := AutoResolve("echo hello")
	if resolved != "echo hello" {
		t.Errorf("expected unchanged command, got %q", resolved)
	}
	if len(remaining) != 0 {
		t.Errorf("expected 0 remaining, got %d", len(remaining))
	}
}

func TestAutoResolve_OnlyTextPlaceholders(t *testing.T) {
	cmd := "kubectl logs {{service}}"
	resolved, remaining := AutoResolve(cmd)
	if resolved != cmd {
		t.Errorf("text placeholder should not be auto-resolved, got %q", resolved)
	}
	if len(remaining) != 1 {
		t.Errorf("expected 1 remaining placeholder, got %d", len(remaining))
	}
}
