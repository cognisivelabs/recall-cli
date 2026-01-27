package placeholders

import (
	"regexp"
	"strings"
)

var (
	// Matches {{name}} or {{options:a,b,c}}
	re = regexp.MustCompile(`\{\{([^}]+)\}\}`)
)

type Placeholder struct {
	FullMatch string // {{name}}
	Key       string // name
	Type      string // "text" or "options"
	Options   []string
}

func Parse(command string) []Placeholder {
	matches := re.FindAllStringSubmatch(command, -1)
	var placeholders []Placeholder

	for _, match := range matches {
		full := match[0]    // {{...}}
		content := match[1] // internal text

		p := Placeholder{
			FullMatch: full,
		}

		if strings.HasPrefix(content, "options:") {
			p.Type = "options"
			p.Key = "Select option"
			optsPart := strings.TrimPrefix(content, "options:")
			p.Options = strings.Split(optsPart, ",")
		} else {
			p.Type = "text"
			p.Key = content // e.g. "branch" or "file"
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
