// Package knowledge says which files of a project belong to its knowledge base.
package knowledge

import "strings"

var folders = []string{".claude/docs/", ".claude/rules/", ".claude/agents/", ".claude/skills/"}

// Covers reports whether a path, relative to the project root and with forward slashes, is part
// of the knowledge base: the router, the docs, rules, agents and skills, and every deltas/ folder.
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
	return len(parts) > 1 && (parts[0] == "deltas" || (len(parts) > 2 && parts[1] == "deltas"))
}
