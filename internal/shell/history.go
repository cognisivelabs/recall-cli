package shell

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetLastCommand reads the last command from the user's history file.
// Currently supports Zsh (~/.zsh_history).
func GetLastCommand() (string, error) {
	// TODO: Detect shell and read appropriate file
	// For MVP, assuming Zsh on Mac as per prompt
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	historyFile := filepath.Join(home, ".zsh_history")

	// Open file
	file, err := os.Open(historyFile)
	if err != nil {
		// Fallback to bash if zsh not found
		historyFile = filepath.Join(home, ".bash_history")
		file, err = os.Open(historyFile)
		if err != nil {
			return "", fmt.Errorf("could not open history file: %w", err)
		}
	}
	defer file.Close()

	// Read last line efficiently?
	// History files can be large, so seeking to end might be better,
	// but lines are variable length.
	// For MVP, just scan.
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

	// Zsh history format: ": 1678900000:0;command"
	// We need to parse this.
	return parseZshLine(lastLine), nil
}

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
