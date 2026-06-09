package shell

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/lipgloss"
)

var cmdPreview = lipgloss.NewStyle().Faint(true)

// Execute runs a command string via the user's shell.
// It prints a dim preview of the command to stderr before running.
// Returns the process exit code (0 = success) and any execution error.
func Execute(command string) (int, error) {
	fmt.Fprintln(os.Stderr, cmdPreview.Render("$ "+command))

	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "/bin/sh"
	}
	cmd := exec.Command(sh, "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return -1, err
	}
	return 0, nil
}
