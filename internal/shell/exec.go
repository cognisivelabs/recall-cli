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
func Execute(command string) {
	fmt.Fprintln(os.Stderr, cmdPreview.Render("$ "+command))

	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "/bin/sh"
	}
	c := exec.Command(sh, "-c", command)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Run()
}
