package spec

import (
	"fmt"
	"strings"

	"github.com/ronz204/deltas-ai/internal/domain/diag"
)

// Parse reads a spec. It reports every problem it finds, not just the first, and Claims holds
// only the claims that parsed cleanly, so check the diagnostics before trusting the result.
// name is only used to label diagnostics.
//
// shortcut: a list item whose first word looks like an ID (letters, a hyphen, digits) is taken
// as a claim attempt, so a prose bullet like "- ISO-8601 dates" is reported; reword it.
func Parse(name string, src []byte) (Spec, []diag.Diagnostic) {
	p := &parser{name: name}
	lines := splitLines(src)
	var sp Spec
	if body, ok := p.frontmatter(lines, &sp); ok {
		sp.Claims = p.claims(lines, body)
	}
	return sp, p.diags
}

// parser collects the diagnostics of one Parse.
type parser struct {
	name  string
	diags []diag.Diagnostic
}

func (p *parser) add(line int, code, format string, args ...any) {
	p.diags = append(p.diags, diag.Diagnostic{File: p.name, Line: line, Code: code, Message: fmt.Sprintf(format, args...)})
}

// splitLines drops a leading byte order mark and the CR of CRLF endings, so a spec parses the
// same whichever way the platform checked it out.
func splitLines(src []byte) []string {
	lines := strings.Split(strings.TrimPrefix(string(src), "\xef\xbb\xbf"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	return lines
}
