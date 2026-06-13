package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// GetLastCommand returns the most recent command from the user's shell history.
// Checks $HISTFILE first, then ~/.zsh_history, then ~/.bash_history.
// Returns an error if no readable history file is found.
func GetLastCommand() (string, error) {
	f, err := openHistoryFile()
	if err != nil {
		return "", err
	}
	defer f.Close()
	return readLastCommand(f)
}

// openHistoryFile returns an open handle to the first readable history file.
// Resolution order: $HISTFILE → ~/.zsh_history → ~/.bash_history.
func openHistoryFile() (*os.File, error) {
	candidates, err := historyFileCandidates()
	if err != nil {
		return nil, err
	}
	for _, path := range candidates {
		if f, err := os.Open(path); err == nil {
			return f, nil
		}
	}
	return nil, fmt.Errorf("no readable shell history file found (tried: %s)", strings.Join(candidates, ", "))
}

// historyFileCandidates returns the ordered list of paths to try.
func historyFileCandidates() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var candidates []string
	if h := os.Getenv("HISTFILE"); h != "" {
		candidates = append(candidates, h)
	}
	candidates = append(candidates,
		filepath.Join(home, ".zsh_history"),
		filepath.Join(home, ".bash_history"),
	)
	return candidates, nil
}

// readLastCommand reads r line by line and returns the last non-empty line,
// with zsh extended-history prefix stripped.
// Extracted so it can be unit-tested without touching the filesystem.
func readLastCommand(r io.Reader) (string, error) {
	var lastLine string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			lastLine = line
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return parseHistoryLine(lastLine), nil
}

// parseHistoryLine strips the zsh extended-history prefix from a line.
// Zsh extended format: ": <timestamp>:<elapsed>;<command>"
// Plain lines (bash or zsh without EXTENDED_HISTORY) are returned unchanged.
func parseHistoryLine(line string) string {
	if strings.HasPrefix(line, ":") {
		if parts := strings.SplitN(line, ";", 2); len(parts) == 2 {
			return parts[1]
		}
	}
	return line
}
