package lint

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ronz204/deltas-ai/internal/domain/claim"
	"github.com/ronz204/deltas-ai/internal/domain/diag"
	"github.com/ronz204/deltas-ai/internal/domain/spec"
)

// Diagnostic codes. Hooks and CI depend on these, so they do not change once released.
const (
	CodeNoCite    = "LINT-NO-CITE"    // a test claim that no test file under its command cites
	CodeUnknownID = "LINT-UNKNOWN-ID" // a test file cites an ID no spec defines
	CodeSkip      = "LINT-SKIP"       // a file that cites IDs skips or focuses a test
)

// Slice is a parsed spec and the file it came from.
type Slice struct {
	File string
	Spec spec.Spec
}

var (
	looseID  = regexp.MustCompile(`(?i)(?:INV|ACC)[-_]?[0-9]+`)
	strictID = regexp.MustCompile(`(INV|ACC)[-_]?([0-9]+)`) // upper case only: `inv1` is usually a variable
	skipped  = regexp.MustCompile(`\bt\.Skip(?:f|Now)?\(|\.(?:skip|only)\(|\bx(?:it|describe|test)\(|pytest\.mark\.skip|pytest\.skip\(|unittest\.skip`)
)

// Check cross-checks specs against test files. Pass only specs that parsed without problems: a
// broken contract cannot be crossed, and its dropped claims would show up as unknown IDs.
// The diagnostics come in a fixed order, claims first and then files in the order given.
func Check(slices []Slice, files []File) []diag.Diagnostic {
	var out []diag.Diagnostic
	defined := map[string]bool{}
	for _, s := range slices {
		for _, c := range s.Spec.Claims {
			defined[c.ID.String()] = true
			if c.Tag.Kind == claim.Test && !Cited(c.ID, c.Tag.Arg, files) {
				out = append(out, diag.Diagnostic{File: s.File, Line: c.Line, Code: CodeNoCite,
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

// unknownIDs warns, once per ID, about each ID the file cites that no spec defines.
func unknownIDs(file string, lines []string, defined map[string]bool) []diag.Diagnostic {
	var out []diag.Diagnostic
	reported := map[string]bool{}
	for i, text := range lines {
		for _, m := range strictID.FindAllStringSubmatch(text, -1) {
			id := m[1] + "-" + m[2]
			if defined[id] || reported[id] {
				continue
			}
			reported[id] = true
			out = append(out, diag.Diagnostic{File: file, Line: i + 1, Code: CodeUnknownID, Warn: true,
				Message: fmt.Sprintf("cites %s, which no spec defines", id)})
		}
	}
	return out
}

// skippedTests warns about the first skipped or focused test in the file, if any.
//
// shortcut: skipped tests are found by pattern and flagged per file, not per test; a file that
// cites IDs and skips an unrelated test is warned too. It is a warning, not a verdict.
func skippedTests(file string, lines []string) []diag.Diagnostic {
	for i, text := range lines {
		if found := skipped.FindString(text); found != "" {
			return []diag.Diagnostic{{File: file, Line: i + 1, Code: CodeSkip, Warn: true,
				Message: fmt.Sprintf("this file cites claim IDs and has a skipped or focused test (%s); a skipped test passes without running", found)}}
		}
	}
	return nil
}
