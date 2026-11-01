// Package commits checks a commit message against the project's Conventional Commits rules.
package commits

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ronz204/deltas-ai/sources/domain/findings"
)

// Diagnostic codes. Hooks and CI depend on these, so they do not change once released.
const (
	CodeEmpty  = "COMMIT-EMPTY"
	CodeFormat = "COMMIT-FORMAT"
	CodeType   = "COMMIT-TYPE"
	CodeLength = "COMMIT-LENGTH"
	CodeCase   = "COMMIT-CASE"
	CodePeriod = "COMMIT-PERIOD"
	CodeBody   = "COMMIT-BODY"
)

const (
	minLength = 40
	maxLength = 50
)

var types = []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"}

var subjectLine = regexp.MustCompile(`^([a-z]+)(?:\([^()\s]+\))?: (.+)$`)

// generated are the prefixes of messages git writes itself; they are not held to the format.
var generated = []string{"Merge ", `Revert "`, "fixup! ", "squash! ", "amend! "}

type line struct {
	n    int
	text string
}

// shortcut: the imperative mood is not checked; no rule tells "embed" from "embedded" without a dictionary.
func Validate(name, message string) []findings.Diagnostic {
	lines := significant(message)
	if len(lines) == 0 {
		return []findings.Diagnostic{{File: name, Code: CodeEmpty, Message: "the commit message is empty"}}
	}
	subject := lines[0]
	for _, prefix := range generated {
		if strings.HasPrefix(subject.text, prefix) {
			return nil
		}
	}
	out := checkSubject(name, subject)
	if len(lines) > 1 && lines[1].n == subject.n+1 && strings.TrimSpace(lines[1].text) != "" {
		out = append(out, findings.Diagnostic{File: name, Line: lines[1].n, Code: CodeBody, Message: "separate the body from the subject with a blank line"})
	}
	return out
}

// significant drops comment lines and leading blanks; inner blanks stay, they separate subject from body.
func significant(message string) []line {
	var out []line
	for i, text := range strings.Split(strings.ReplaceAll(message, "\r\n", "\n"), "\n") {
		if strings.HasPrefix(text, "#") || (len(out) == 0 && strings.TrimSpace(text) == "") {
			continue
		}
		out = append(out, line{n: i + 1, text: text})
	}
	return out
}

func checkSubject(name string, subject line) []findings.Diagnostic {
	var out []findings.Diagnostic
	add := func(code, format string, args ...any) {
		out = append(out, findings.Diagnostic{File: name, Line: subject.n, Code: code, Message: fmt.Sprintf(format, args...)})
	}
	if n := utf8.RuneCountInString(subject.text); n < minLength || n > maxLength {
		add(CodeLength, "the subject is %d characters; it must be %d to %d, type and scope included", n, minLength, maxLength)
	}
	m := subjectLine.FindStringSubmatch(subject.text)
	if m == nil {
		add(CodeFormat, "the subject must read type(scope): subject")
		return out
	}
	kind, description := m[1], m[2]
	if !slices.Contains(types, kind) {
		add(CodeType, "type %q is not allowed; use %s", kind, strings.Join(types, ", "))
	}
	if first, _ := utf8.DecodeRuneInString(description); unicode.IsUpper(first) {
		add(CodeCase, "the subject starts lowercase after the colon")
	}
	if strings.HasSuffix(description, ".") {
		add(CodePeriod, "the subject has no trailing period")
	}
	return out
}
