package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- parseHistoryLine ---

func TestParseHistoryLine_ZshExtended(t *testing.T) {
	line := ": 1678900000:0;kubectl get pods"
	got := parseHistoryLine(line)
	if got != "kubectl get pods" {
		t.Errorf("expected 'kubectl get pods', got %q", got)
	}
}

func TestParseHistoryLine_ZshExtended_CommandWithSemicolon(t *testing.T) {
	// The command itself contains a semicolon — only the first one is the separator.
	line := ": 1678900000:0;echo a; echo b"
	got := parseHistoryLine(line)
	if got != "echo a; echo b" {
		t.Errorf("expected 'echo a; echo b', got %q", got)
	}
}

func TestParseHistoryLine_BashPlain(t *testing.T) {
	line := "docker compose up -d"
	got := parseHistoryLine(line)
	if got != "docker compose up -d" {
		t.Errorf("expected plain bash line unchanged, got %q", got)
	}
}

func TestParseHistoryLine_Empty(t *testing.T) {
	got := parseHistoryLine("")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestParseHistoryLine_ColonButNoSemicolon(t *testing.T) {
	// Starts with colon but is not a valid zsh extended line — return as-is.
	line := ":not-a-zsh-line"
	got := parseHistoryLine(line)
	if got != ":not-a-zsh-line" {
		t.Errorf("expected line unchanged, got %q", got)
	}
}

// --- readLastCommand ---

func TestReadLastCommand_ZshHistory(t *testing.T) {
	input := strings.NewReader(": 1678900000:0;kubectl get pods\n: 1678900001:0;docker ps\n")
	got, err := readLastCommand(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "docker ps" {
		t.Errorf("expected 'docker ps', got %q", got)
	}
}

func TestReadLastCommand_BashHistory(t *testing.T) {
	input := strings.NewReader("git status\ngit diff\ngit commit -m 'fix'\n")
	got, err := readLastCommand(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "git commit -m 'fix'" {
		t.Errorf("expected last git command, got %q", got)
	}
}

func TestReadLastCommand_SkipsBlankLines(t *testing.T) {
	input := strings.NewReader("echo hello\n\n   \n")
	got, err := readLastCommand(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "echo hello" {
		t.Errorf("expected 'echo hello', got %q", got)
	}
}

func TestReadLastCommand_Empty(t *testing.T) {
	input := strings.NewReader("")
	got, err := readLastCommand(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string for empty history, got %q", got)
	}
}

func TestReadLastCommand_SingleLine(t *testing.T) {
	input := strings.NewReader("ls -la\n")
	got, err := readLastCommand(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ls -la" {
		t.Errorf("expected 'ls -la', got %q", got)
	}
}

// --- historyFileCandidates ---

func TestHistoryFileCandidates_IncludesHISTFILE(t *testing.T) {
	prev, existed := os.LookupEnv("HISTFILE")
	os.Setenv("HISTFILE", "/custom/history")
	t.Cleanup(func() {
		if existed {
			os.Setenv("HISTFILE", prev)
		} else {
			os.Unsetenv("HISTFILE")
		}
	})

	candidates, err := historyFileCandidates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if candidates[0] != "/custom/history" {
		t.Errorf("expected HISTFILE first, got %q", candidates[0])
	}
}

func TestHistoryFileCandidates_DefaultsWithoutHISTFILE(t *testing.T) {
	prev, existed := os.LookupEnv("HISTFILE")
	os.Unsetenv("HISTFILE")
	t.Cleanup(func() {
		if existed {
			os.Setenv("HISTFILE", prev)
		}
	})

	candidates, err := historyFileCandidates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(candidates) < 2 {
		t.Fatalf("expected at least 2 default candidates, got %d", len(candidates))
	}

	home, _ := os.UserHomeDir()
	if candidates[0] != filepath.Join(home, ".zsh_history") {
		t.Errorf("expected zsh_history first, got %q", candidates[0])
	}
	if candidates[1] != filepath.Join(home, ".bash_history") {
		t.Errorf("expected bash_history second, got %q", candidates[1])
	}
}

// --- GetLastCommand (integration: reads a real temp file) ---

func TestGetLastCommand_ReadsFile(t *testing.T) {
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, "history")
	os.WriteFile(histFile, []byte(": 1678900000:0;make build\n: 1678900001:0;make test\n"), 0644)

	prev, existed := os.LookupEnv("HISTFILE")
	os.Setenv("HISTFILE", histFile)
	t.Cleanup(func() {
		if existed {
			os.Setenv("HISTFILE", prev)
		} else {
			os.Unsetenv("HISTFILE")
		}
	})

	got, err := GetLastCommand()
	if err != nil {
		t.Fatalf("GetLastCommand: %v", err)
	}
	if got != "make test" {
		t.Errorf("expected 'make test', got %q", got)
	}
}

func TestGetLastCommand_ErrorWhenNoFile(t *testing.T) {
	// Point all candidates at non-existent paths.
	tmpDir := t.TempDir()

	for _, key := range []string{"HISTFILE"} {
		prev, existed := os.LookupEnv(key)
		os.Setenv(key, filepath.Join(tmpDir, "no-such-file"))
		t.Cleanup(func() {
			if existed {
				os.Setenv(key, prev)
			} else {
				os.Unsetenv(key)
			}
		})
	}

	// Also redirect HOME so the default ~/.zsh_history / ~/.bash_history don't exist.
	prev, existed := os.LookupEnv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		if existed {
			os.Setenv("HOME", prev)
		} else {
			os.Unsetenv("HOME")
		}
	})

	_, err := GetLastCommand()
	if err == nil {
		t.Fatal("expected error when no history file exists, got nil")
	}
}
