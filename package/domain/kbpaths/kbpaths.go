package kbpaths

import (
	"slices"
	"strings"
)

var folders = []string{".claude/docs/", ".claude/rules/", ".claude/agents/", ".claude/skills/"}

func Covers(path string) bool {
	if path == "CLAUDE.md" || path == ".claude/CLAUDE.md" {
		return true
	}
	if slices.ContainsFunc(folders, func(folder string) bool { return strings.HasPrefix(path, folder) }) {
		return true
	}
	parts := strings.Split(path, "/")
	n := len(parts)
	inDeltas := n > 1 && parts[n-2] == "deltas"
	inSpecs := n > 2 && parts[n-2] == "specs" && parts[n-3] == "deltas"
	return inDeltas || inSpecs
}
