package workspace

import (
	"os"
	"path/filepath"
	"strings"
)

// Detect returns the absolute path of the current working directory.
// Used to decide which saved commands are relevant to the user's current context.
func Detect() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}

// Matches reports whether cwd satisfies filter.
// filter can be a glob pattern (e.g. "~/work/billing-*") or a path prefix.
// Returns false when either argument is empty.
func Matches(cwd, filter string) bool {
	if filter == "" || cwd == "" {
		return false
	}

	filter = expandHome(filter)

	if matched, _ := filepath.Match(filter, cwd); matched {
		return true
	}

	return strings.HasPrefix(cwd, filter)
}

// expandHome replaces a leading "~/" with the user's home directory.
func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
