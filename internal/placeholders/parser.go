package placeholders

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	re = regexp.MustCompile(`\{\{([^}]+)\}\}`)

	// Placeholders that resolve automatically without user input
	autoResolvers = map[string]func() string{
		"branch": gitBranch,
		"cwd":    cwd,
		"dir":    dirName,
		"user":   userName,
		"host":   hostName,
		"home":   homeDir,
	}
)

type Placeholder struct {
	FullMatch string
	Key       string
	Type      string // "text", "options", or "auto"
	Options   []string
	AutoValue string // pre-resolved value for auto type
}

// Parse finds all {{...}} tokens in command and returns a Placeholder for each.
// Token types:
//   - "options:a,b,c"  → user picks from a list
//   - known key (branch, cwd, …) → resolved automatically at runtime
//   - anything else    → user types a free-form value
func Parse(command string) []Placeholder {
	matches := re.FindAllStringSubmatch(command, -1)
	var placeholders []Placeholder

	for _, match := range matches {
		full := match[0]
		content := match[1]

		p := Placeholder{
			FullMatch: full,
		}

		if strings.HasPrefix(content, "options:") {
			p.Type = "options"
			p.Key = "Select option"
			optsPart := strings.TrimPrefix(content, "options:")
			p.Options = strings.Split(optsPart, ",")
		} else if resolver, ok := autoResolvers[content]; ok {
			p.Type = "auto"
			p.Key = content
			p.AutoValue = resolver()
		} else {
			p.Type = "text"
			p.Key = content
		}

		placeholders = append(placeholders, p)
	}

	return placeholders
}

// HasPlaceholders reports whether command contains any {{...}} tokens.
func HasPlaceholders(command string) bool {
	return re.MatchString(command)
}

// Replace substitutes the first occurrence of p.FullMatch in command with value.
func Replace(command string, p Placeholder, value string) string {
	return strings.Replace(command, p.FullMatch, value, 1)
}

// AutoResolve replaces all "auto" type placeholders (branch, cwd, user, etc.) in one pass.
// Returns the updated command string and any placeholders that still need user input.
func AutoResolve(command string) (string, []Placeholder) {
	all := Parse(command)
	var remaining []Placeholder

	for _, p := range all {
		if p.Type == "auto" && p.AutoValue != "" {
			command = Replace(command, p, p.AutoValue)
		} else {
			remaining = append(remaining, p)
		}
	}

	return command, remaining
}

// gitBranch returns the current git branch name, or "" if not in a git repo.
func gitBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// cwd returns the absolute path of the current working directory.
func cwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

// dirName returns just the last segment of the current working directory path.
func dirName() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(dir)
}

// userName returns the current OS username.
func userName() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.Username
}

// hostName returns the machine hostname.
func hostName() string {
	h, err := os.Hostname()
	if err != nil {
		return ""
	}
	return h
}

// homeDir returns the current user's home directory.
func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}
