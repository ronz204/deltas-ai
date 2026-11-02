package specfile

import (
	"fmt"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/findings"
)

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

func splitLines(src []byte) []string {
	lines := strings.Split(strings.TrimPrefix(string(src), "\xef\xbb\xbf"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	return lines
}
