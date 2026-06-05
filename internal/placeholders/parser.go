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

func HasPlaceholders(command string) bool {
	return re.MatchString(command)
}

func Replace(command string, p Placeholder, value string) string {
	return strings.Replace(command, p.FullMatch, value, 1)
}

// AutoResolve replaces all auto-resolvable placeholders in one pass.
// Returns the command and the remaining placeholders that need user input.
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

func gitBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func cwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

func dirName() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(dir)
}

func userName() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.Username
}

func hostName() string {
	h, err := os.Hostname()
	if err != nil {
		return ""
	}
	return h
}

func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}
