// Package knowledge says which files of a project belong to its knowledge base.
package knowledge

import "strings"

var folders = []string{".claude/docs/", ".claude/rules/", ".claude/agents/", ".claude/skills/"}

// Covers reports whether a path, relative to the project root and with forward slashes, is part
// of the knowledge base: the router, the docs, rules, agents and skills, and every deltas/ folder (its config, its manifest and its specifies/).
func Covers(path string) bool {
	if path == "CLAUDE.md" || path == ".claude/CLAUDE.md" {
		return true
	}
	for _, folder := range folders {
		if strings.HasPrefix(path, folder) {
			return true
		}
	}
	parts := strings.Split(path, "/")
	n := len(parts)
	if n > 1 && parts[n-2] == "deltas" {
		return true
	}
	return n > 2 && parts[n-2] == "specifies" && parts[n-3] == "deltas"
}
