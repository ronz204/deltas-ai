package guides

import (
	"path"
	"strings"
)

const Dir = "deltas/docs"

var Order = []string{"overview", "structure", "expertise", "approach"}

func Name(p string) (string, bool) {
	base, ok := strings.CutSuffix(strings.TrimPrefix(p, Dir+"/"), ".md")
	return base, ok && strings.HasPrefix(p, Dir+"/") && path.Base(base) == base
}

func Filled(text string) bool { return !strings.Contains(text, "<!--") }
