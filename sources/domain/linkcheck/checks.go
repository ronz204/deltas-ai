package linkcheck

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ronz204/deltas-ai/sources/domain/assertion"
	"github.com/ronz204/deltas-ai/sources/domain/findings"
	"github.com/ronz204/deltas-ai/sources/domain/specfile"
)

// Diagnostic codes. Hooks and CI depend on these, so they do not change once released.
const (
	CodeNoCite    = "LINT-NO-CITE"
	CodeUnknownID = "LINT-UNKNOWN-ID"
	CodeSkip      = "LINT-SKIP"
)

type Slice struct {
	File string
	Spec specfile.Spec
}

var (
	looseID  = regexp.MustCompile(`(?i)(?:INV|ACC)[-_]?[0-9]+`)
	strictID = regexp.MustCompile(`(INV|ACC)[-_]?([0-9]+)`) // upper case only: `inv1` is usually a variable
	skipped  = regexp.MustCompile(`\bt\.Skip(?:f|Now)?\(|\.(?:skip|only)\(|\bx(?:it|describe|test)\(|pytest\.mark\.skip|pytest\.skip\(|unittest\.skip`)
)

// Check needs specs that parsed cleanly: a dropped claim would show up as an unknown ID.
func Check(slices []Slice, files []File) []findings.Diagnostic {
	var out []findings.Diagnostic
	defined := map[string]bool{}
	for _, s := range slices {
		for _, c := range s.Spec.Claims {
			defined[c.ID.String()] = true
			if c.Tag.Kind == assertion.Test && !Cited(c.ID, c.Tag.Arg, files) {
				out = append(out, findings.Diagnostic{File: s.File, Line: c.Line, Code: CodeNoCite,
					Message: fmt.Sprintf("%s has a test tag, but no test file under %q cites its ID", c.ID, c.Tag.Arg)})
			}
		}
	}
	for _, f := range files {
		lines := strings.Split(f.Text, "\n")
		out = append(out, unknownIDs(f.Path, lines, defined)...)
		if looseID.MatchString(f.Text) {
			out = append(out, skippedTests(f.Path, lines)...)
		}
	}
	return out
}

func unknownIDs(file string, lines []string, defined map[string]bool) []findings.Diagnostic {
	var out []findings.Diagnostic
	reported := map[string]bool{}
	for i, text := range lines {
		for _, m := range strictID.FindAllStringSubmatch(text, -1) {
			id := m[1] + "-" + m[2]
			if defined[id] || reported[id] {
				continue
			}
			reported[id] = true
			out = append(out, findings.Diagnostic{File: file, Line: i + 1, Code: CodeUnknownID, Warn: true,
				Message: fmt.Sprintf("cites %s, which no spec defines", id)})
		}
	}
	return out
}

// shortcut: skipped tests are found by pattern and flagged per file; it is a warning, not a verdict.
func skippedTests(file string, lines []string) []findings.Diagnostic {
	for i, text := range lines {
		if found := skipped.FindString(text); found != "" {
			return []findings.Diagnostic{{File: file, Line: i + 1, Code: CodeSkip, Warn: true,
				Message: fmt.Sprintf("this file cites claim IDs and has a skipped or focused test (%s); a skipped test passes without running", found)}}
		}
	}
	return nil
}
