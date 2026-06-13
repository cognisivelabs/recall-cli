package shell

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetLastCommand returns the most recent command from the user's shell history.
// Checks ~/.zsh_history first (zsh extended format), then falls back to ~/.bash_history.
// Returns an error if neither file is readable.
func GetLastCommand() (string, error) {
	// TODO: Detect shell and read appropriate file
	// For MVP, assuming Zsh on Mac as per prompt
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	historyFile := filepath.Join(home, ".zsh_history")

	file, err := os.Open(historyFile)
	if err != nil {
		historyFile = filepath.Join(home, ".bash_history")
		file, err = os.Open(historyFile)
		if err != nil {
			return "", fmt.Errorf("could not open history file: %w", err)
		}
	}
	defer file.Close()

	var lastLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			lastLine = line
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return parseZshLine(lastLine), nil
}

// parseZshLine strips the zsh extended-history prefix from a line.
// Zsh extended format is ": <timestamp>:<elapsed>;<command>".
// Plain lines (bash or zsh without EXTENDED_HISTORY) are returned unchanged.
func parseZshLine(line string) string {
	// Zsh extended history format
	if strings.HasPrefix(line, ":") {
		parts := strings.SplitN(line, ";", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return line
}
