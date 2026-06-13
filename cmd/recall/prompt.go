package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
)

// confirmFromReader reads a y/N confirmation from the given reader.
// Uses bufio to handle piped input cleanly (echo y | recall add ...).
// Outputs the prompt to the given writer (typically stderr).
// Used by: add, delete, save commands.
func confirmFromReader(in io.Reader, errOut io.Writer, prompt string) bool {
	_, _ = fmt.Fprint(errOut, prompt)
	reader := bufio.NewReader(in)
	line, _ := reader.ReadString('\n')
	input := strings.ToLower(strings.TrimSpace(line))
	return input == "y" || input == "yes"
}

// printMatchList prints a numbered list of matching commands to the given writer.
// Used when a query matches multiple commands and the user needs to pick one.
// Used by: delete, run commands.
func printMatchList(out io.Writer, query string, matches []storage.Command) {
	_, _ = fmt.Fprintf(out, "Multiple matches for %q:\n", query)
	for i, cmd := range matches {
		_, _ = fmt.Fprintf(out, "  [%d] %s — %s\n", i+1, cmd.Pattern, cmd.Description)
	}
}

// findCommandByIDOrPattern resolves a command from a numeric ID or an exact pattern string.
// Returns an error if the argument is empty or no match is found.
// Used by: edit, delete commands.
func findCommandByIDOrPattern(store storage.Storage, arg string) (*storage.Command, error) {
	if arg == "" {
		return nil, fmt.Errorf("argument cannot be empty")
	}

	if id, err := strconv.Atoi(arg); err == nil {
		target, err := store.GetByID(id)
		if err != nil {
			return nil, fmt.Errorf("looking up command: %w", err)
		}
		if target == nil {
			return nil, fmt.Errorf("no command with ID %d found", id)
		}
		return target, nil
	}

	target, err := store.GetByPattern(arg)
	if err != nil {
		return nil, fmt.Errorf("looking up command: %w", err)
	}
	if target == nil {
		return nil, fmt.Errorf("no command matching %q found", arg)
	}
	return target, nil
}
