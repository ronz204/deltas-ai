package linking

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/claimset"
	"github.com/ronz204/deltas-ai/package/domain/findings"
	"github.com/ronz204/deltas-ai/package/domain/specfile"
)

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
	strictID = regexp.MustCompile(`(INV|ACC)[-_]?([0-9]+)`)
	skipped  = regexp.MustCompile(`\bt\.Skip(?:f|Now)?\(|\.(?:skip|only)\(|\bx(?:it|describe|test)\(|pytest\.mark\.skip|pytest\.skip\(|unittest\.skip`)
)

func Check(slices []Slice, files []File) []findings.Diagnostic {
	var out []findings.Diagnostic
	defined := map[string]bool{}
	for _, s := range slices {
		for _, c := range s.Spec.Claims {
			defined[c.ID.String()] = true
			if c.Tag.Method == claimset.Test && !Cited(c.ID, c.Tag.Arg, files) {
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

func skippedTests(file string, lines []string) []findings.Diagnostic {
	for i, text := range lines {
		if found := skipped.FindString(text); found != "" {
			return []findings.Diagnostic{{File: file, Line: i + 1, Code: CodeSkip, Warn: true,
				Message: fmt.Sprintf("this file cites claim IDs and has a skipped or focused test (%s); a skipped test passes without running", found)}}
		}
	}
	return nil
}
