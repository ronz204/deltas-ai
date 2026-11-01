package specfile

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/ronz204/deltas-ai/internal/domain/lifecycle"
)

var keyLine = regexp.MustCompile(`^([a-z_]+):\s*(.*?)\s*$`)

var knownKeys = map[string]bool{"format": true, "slice": true, "status": true}

type entry struct {
	value string
	line  int
}

// frontmatter reads the top block into sp; ok is false when there is none, and then nothing else is read.
func (p *parser) frontmatter(lines []string, sp *Spec) (body int, ok bool) {
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

// keys reads "key: value" lines. firstLine is the line number of block[0].
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

// fill checks each key and sets the field; a bad value leaves the field empty.
func (p *parser) fill(kv map[string]entry, sp *Spec) {
	if e, ok := p.need(kv, "format"); ok {
		sp.Format = p.format(e)
	}
	if e, ok := p.need(kv, "slice"); ok {
		sp.Slice = p.sliceOf(e)
	}
	if e, ok := p.need(kv, "status"); ok {
		sp.Status = p.statusOf(e)
	}
}

func (p *parser) need(kv map[string]entry, key string) (entry, bool) {
	e, ok := kv[key]
	if !ok {
		p.add(1, CodeKey, "the frontmatter is missing %q", key)
	}
	return e, ok
}

func (p *parser) format(e entry) int {
	n, err := strconv.Atoi(e.value)
	if err != nil || n != supportedFormat {
		p.add(e.line, CodeFormat, "format %q is not supported; this version reads format %d", e.value, supportedFormat)
		return 0
	}
	return n
}

func (p *parser) sliceOf(e entry) string {
	if !ValidSlice(e.value) {
		p.add(e.line, CodeSlice, "slice %q must be lowercase words joined by hyphens", e.value)
		return ""
	}
	return e.value
}

func (p *parser) statusOf(e entry) lifecycle.Status {
	st, err := lifecycle.ParseStatus(e.value)
	if err != nil {
		p.add(e.line, CodeStatus, "%v", err)
	}
	return st
}
