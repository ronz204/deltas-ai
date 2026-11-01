package specfile

import (
	"regexp"
	"strings"

	"github.com/ronz204/deltas-ai/source/domain/assertion"
)

var (
	headerLike = regexp.MustCompile(`^- [A-Za-z]+-[0-9]+(?:\s|$)`)
	headerLine = regexp.MustCompile(`^- ([A-Za-z]+-[0-9]+)\s+\[(.*)\]\s*$`)
)

// claims skips fenced code and HTML comments, so a spec can show an example claim without declaring it.
func (p *parser) claims(lines []string, from int) []assertion.Claim {
	var out []assertion.Claim
	seen := map[assertion.ID]int{} // each ID with the line that defined it
	var skip skipper
	for i := from; i < len(lines); i++ {
		if skip.skips(lines[i]) || !headerLike.MatchString(lines[i]) {
			continue
		}
		text, last := continuation(lines, i+1)
		if c, ok := p.claim(i+1, lines[i], text, seen); ok {
			out = append(out, c)
		}
		i = last
	}
	return out
}

// claim checks one header and its sentence. ok is false when anything about it was reported.
func (p *parser) claim(line int, header, text string, seen map[assertion.ID]int) (c assertion.Claim, ok bool) {
	m := headerLine.FindStringSubmatch(header)
	if m == nil {
		p.add(line, CodeClaim, `a claim header reads "- ID [tag]"`)
		return assertion.Claim{}, false
	}
	id, err := assertion.ParseID(m[1])
	if err != nil {
		p.add(line, CodeClaim, "%v", err)
		return assertion.Claim{}, false
	}
	if first, dup := seen[id]; dup {
		p.add(line, CodeDupID, "%s is already defined at line %d", id, first)
		return assertion.Claim{}, false
	}
	seen[id] = line

	before := len(p.diags)
	tag, err := assertion.ParseTag(m[2])
	if err != nil {
		p.add(line, CodeTag, "%s: %v", id, err)
	}
	if text == "" {
		p.add(line, CodeNoText, "%s has no sentence; indent it under the header", id)
	} else if err := assertion.CheckStatement(id.Kind(), text); err != nil {
		p.add(line, CodeStatement, "%s: %v", id, err)
	}
	return assertion.Claim{ID: id, Tag: tag, Text: text, Line: line}, len(p.diags) == before
}

// continuation joins the indented lines under a header; last is the index of the final line used.
func continuation(lines []string, from int) (text string, last int) {
	var parts []string
	last = from - 1
	for _, l := range lines[from:] {
		if strings.TrimSpace(l) == "" || (l[0] != ' ' && l[0] != '\t') {
			break
		}
		parts = append(parts, strings.TrimSpace(l))
		last++
	}
	return strings.Join(parts, " "), last
}

type skipper struct{ inFence, inComment bool }

func (s *skipper) skips(line string) bool {
	t := strings.TrimSpace(line)
	switch {
	case s.inComment:
		s.inComment = !strings.Contains(t, "-->")
		return true
	case strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~"):
		s.inFence = !s.inFence
		return true
	case s.inFence:
		return true
	case strings.Contains(t, "<!--"):
		s.inComment = !strings.Contains(t, "-->")
		return true
	}
	return false
}
