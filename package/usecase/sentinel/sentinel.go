package sentinel

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/kbpaths"
	"github.com/ronz204/deltas-ai/package/usecase/tracking"
)

type Guard struct {
	root     string
	slices   *tracking.Reporter
	health   Health
	evidence Evidence
	linter   Linter
}

func New(root string, slices *tracking.Reporter, health Health, evidence Evidence, linter Linter) *Guard {
	return &Guard{root: root, slices: slices, health: health, evidence: evidence, linter: linter}
}

var fileTools = map[string]bool{"Write": true, "Edit": true}

func (g *Guard) Decide(ctx context.Context, e Event) (Decision, error) {
	switch e.Name {
	case "SessionStart":
		return g.start(ctx)
	case "UserPromptSubmit":
		return g.prompt(ctx)
	case "PreToolUse":
		return g.before(e), nil
	case "PostToolUse":
		return g.after(ctx, e)
	case "Stop":
		return g.stop(ctx, e)
	}
	return Decision{}, nil
}

func fileList(files []string) string {
	const shown = 3
	if len(files) <= shown {
		return strings.Join(files, ", ")
	}
	return fmt.Sprintf("%s and %d more", strings.Join(files[:shown], ", "), len(files)-shown)
}

func (g *Guard) before(e Event) Decision {
	if !fileTools[e.Tool] {
		return Decision{}
	}
	if rel, ok := g.relative(e.Path); ok && kbpaths.Covers(rel) {
		return Decision{Block: true, Reason: fmt.Sprintf(
			"%s is part of the knowledge base, which is not edited directly. Specs are written with delta spec write and tests are linked with delta link; a person can edit it by hand.", rel)}
	}
	return Decision{}
}

func (g *Guard) relative(path string) (string, bool) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(g.root, path)
	}
	rel, err := filepath.Rel(g.root, path)
	if err != nil {
		return "", false
	}
	rel = filepath.ToSlash(rel)
	return rel, rel != ".." && !strings.HasPrefix(rel, "../")
}
