package storage

import (
	"sort"
	"strings"
)

// HasTag checks if a comma-separated tag string contains the given tag (case-insensitive).
func HasTag(tags, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		return false
	}
	for _, t := range strings.Split(tags, ",") {
		if strings.ToLower(strings.TrimSpace(t)) == target {
			return true
		}
	}
	return false
}

// CollectTags extracts all unique tags from a list of commands, sorted alphabetically.
func CollectTags(cmds []Command) []string {
	seen := make(map[string]bool)
	var tags []string
	for _, c := range cmds {
		for _, t := range strings.Split(c.Tags, ",") {
			t = strings.TrimSpace(t)
			if t != "" && !seen[t] {
				seen[t] = true
				tags = append(tags, t)
			}
		}
	}
	sort.Strings(tags)
	return tags
}

// FilterByTag returns commands that have the given tag.
func FilterByTag(cmds []Command, tag string) []Command {
	if tag == "" {
		return cmds
	}
	var filtered []Command
	for _, c := range cmds {
		if HasTag(c.Tags, tag) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
