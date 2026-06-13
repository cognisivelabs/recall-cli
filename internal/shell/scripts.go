package shell

import (
	_ "embed"
	"fmt"
	"io"
)

//go:embed scripts/init.zsh
var zshScript string

//go:embed scripts/init.bash
var bashScript string

// PrintInitScript writes the shell integration script for the given shell type to w.
// The scripts are embedded at compile time from the scripts/ directory.
// Returns an error for unsupported shells instead of silently falling back.
func PrintInitScript(w io.Writer, shellType string) error {
	switch shellType {
	case "zsh":
		_, err := fmt.Fprint(w, zshScript)
		return err
	case "bash":
		_, err := fmt.Fprint(w, bashScript)
		return err
	default:
		return fmt.Errorf("unsupported shell: %q (supported: zsh, bash)", shellType)
	}
}
