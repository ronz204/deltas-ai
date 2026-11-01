package spec

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ronz204/deltas-ai/internal/domain/claim"
	"github.com/ronz204/deltas-ai/internal/domain/diag"
	"github.com/ronz204/deltas-ai/internal/domain/slice"
)

var (
	keyLine    = regexp.MustCompile(`^([a-z_]+):\s*(.*?)\s*$`)
	sliceName  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	headerLike = regexp.MustCompile(`^- [A-Za-z]+-[0-9]+(?:\s|$)`)
	headerLine = regexp.MustCompile(`^- ([A-Za-z]+-[0-9]+)\s+\[(.*)\]\s*$`)
)

var knownKeys = map[string]bool{"format": true, "slice": true, "status": true}

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
	body, ok := p.frontmatter(lines, &sp)
	if !ok {
		return sp, p.diags
	}
	sp.Claims = p.claims(lines, body)
	return sp, p.diags
}

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

type entry struct {
	value string
	line  int
}

// frontmatter fills sp and returns the index of the first line after the block.
func (p *parser) frontmatter(lines []string, sp *Spec) (int, bool) {
	if strings.TrimSpace(lines[0]) != "---" {
		p.add(1, CodeFrontmatter, "a spec starts with a --- frontmatter block")
		return 0, false
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		p.add(1, CodeFrontmatter, "the frontmatter is never closed with ---")
		return 0, false
	}
	p.fill(p.keys(lines[1:end], 2), sp)
	return end + 1, true
}

func (p *parser) keys(block []string, firstLine int) map[string]entry {
	kv := map[string]entry{}
	for i, l := range block {
		line := firstLine + i
		if strings.TrimSpace(l) == "" {
			continue
		}
		m := keyLine.FindStringSubmatch(l)
		if m == nil {
			p.add(line, CodeKey, `frontmatter lines look like "key: value"`)
			continue
		}
		key := m[1]
		if !knownKeys[key] {
			p.add(line, CodeKey, "unknown key %q; the keys are format, slice and status", key)
			continue
		}
		if prev, dup := kv[key]; dup {
			p.add(line, CodeKey, "key %q is already set at line %d", key, prev.line)
			continue
		}
		kv[key] = entry{value: m[2], line: line}
	}
	return kv
}

func (p *parser) fill(kv map[string]entry, sp *Spec) {
	if e, ok := p.need(kv, "format"); ok {
		if n, err := strconv.Atoi(e.value); err != nil || n != supportedFormat {
			p.add(e.line, CodeFormat, "format %q is not supported; this version reads format %d", e.value, supportedFormat)
		} else {
			sp.Format = n
		}
	}
	if e, ok := p.need(kv, "slice"); ok {
		if !sliceName.MatchString(e.value) {
			p.add(e.line, CodeSlice, "slice %q must be lowercase words joined by hyphens", e.value)
		} else {
			sp.Slice = e.value
		}
	}
	if e, ok := p.need(kv, "status"); ok {
		if st, err := slice.ParseStatus(e.value); err != nil {
			p.add(e.line, CodeStatus, "%v", err)
		} else {
			sp.Status = st
		}
	}
}

func (p *parser) need(kv map[string]entry, key string) (entry, bool) {
	e, ok := kv[key]
	if !ok {
		p.add(1, CodeKey, "the frontmatter is missing %q", key)
	}
	return e, ok
}

// claims scans the body. Fenced code and HTML comments are skipped, so a spec can show an
// example claim without declaring it.
func (p *parser) claims(lines []string, from int) []claim.Claim {
	var out []claim.Claim
	seen := map[claim.ID]int{}
	var sk skipper
	for i := from; i < len(lines); i++ {
		if sk.skip(lines[i]) || !headerLike.MatchString(lines[i]) {
			continue
		}
		text, next := continuation(lines, i+1)
		if c, ok := p.claim(i+1, lines[i], text, seen); ok {
			out = append(out, c)
		}
		i = next - 1
	}
	return out
}

type skipper struct{ inFence, inComment bool }

func (s *skipper) skip(line string) bool {
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

// continuation gathers the indented lines under a header and returns them joined, plus the
// index of the first line after them.
func continuation(lines []string, from int) (string, int) {
	var parts []string
	i := from
	for ; i < len(lines); i++ {
		l := lines[i]
		if strings.TrimSpace(l) == "" || (l[0] != ' ' && l[0] != '\t') {
			break
		}
		parts = append(parts, strings.TrimSpace(l))
	}
	return strings.Join(parts, " "), i
}

func (p *parser) claim(line int, header, text string, seen map[claim.ID]int) (claim.Claim, bool) {
	m := headerLine.FindStringSubmatch(header)
	if m == nil {
		p.add(line, CodeClaim, `a claim header reads "- ID [tag]"`)
		return claim.Claim{}, false
	}
	id, err := claim.ParseID(m[1])
	if err != nil {
		p.add(line, CodeClaim, "%v", err)
		return claim.Claim{}, false
	}
	if first, dup := seen[id]; dup {
		p.add(line, CodeDupID, "%s is already defined at line %d", id, first)
		return claim.Claim{}, false
	}
	seen[id] = line

	ok := true
	tag, err := claim.ParseTag(m[2])
	if err != nil {
		p.add(line, CodeTag, "%s: %v", id, err)
		ok = false
	}
	if text == "" {
		p.add(line, CodeNoText, "%s has no sentence; indent it under the header", id)
		return claim.Claim{}, false
	}
	if err := claim.CheckStatement(id.Kind(), text); err != nil {
		p.add(line, CodeStatement, "%s: %v", id, err)
		ok = false
	}
	return claim.Claim{ID: id, Tag: tag, Text: text, Line: line}, ok
}
