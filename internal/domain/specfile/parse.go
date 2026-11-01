package specfile

import (
	"fmt"
	"strings"

	"github.com/ronz204/deltas-ai/internal/domain/findings"
)

// Parse reports every problem, and Claims holds only the claims that parsed cleanly: check the diagnostics first.
// shortcut: a list item starting with an ID-like word ("- ISO-8601 dates") is read as a claim attempt.
func Parse(name string, src []byte) (Spec, []findings.Diagnostic) {
	p := &parser{name: name}
	lines := splitLines(src)
	var sp Spec
	if body, ok := p.frontmatter(lines, &sp); ok {
		sp.Claims = p.claims(lines, body)
	}
	return sp, p.diags
}

type parser struct {
	name  string
	diags []findings.Diagnostic
}

func (p *parser) add(line int, code, format string, args ...any) {
	p.diags = append(p.diags, findings.Diagnostic{File: p.name, Line: line, Code: code, Message: fmt.Sprintf(format, args...)})
}

// splitLines drops a BOM and the CR of CRLF, so a spec parses the same on every platform.
func splitLines(src []byte) []string {
	lines := strings.Split(strings.TrimPrefix(string(src), "\xef\xbb\xbf"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	return lines
}
