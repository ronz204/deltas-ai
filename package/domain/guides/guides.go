// Package guides knows the documents written for the people of a project, which delta installs as templates.
package guides

import (
	"path"
	"strings"
)

// Dir is where the docs live in a project.
const Dir = "deltas/docs"

// Order is the order in which the docs are filled in: the vision first, because the others lean on it.
var Order = []string{"overview", "structure", "expertise", "approach"}

// Name is the name of a doc file under Dir without its extension; false when the path is not a doc.
func Name(p string) (string, bool) {
	base, ok := strings.CutSuffix(strings.TrimPrefix(p, Dir+"/"), ".md")
	return base, ok && strings.HasPrefix(p, Dir+"/") && path.Base(base) == base
}

// Filled says whether a doc was written. A template keeps the comments that guide each section and a
// finished doc has deleted them all, so one comment left means the doc is not done.
func Filled(text string) bool { return !strings.Contains(text, "<!--") }
