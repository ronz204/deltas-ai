// Package commit checks a commit message against the project's Conventional Commits rules.
package commit

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ronz204/deltas-ai/internal/domain/diag"
)

// Diagnostic codes. Hooks and CI depend on these, so they do not change once released.
const (
	CodeEmpty  = "COMMIT-EMPTY"  // no subject line
	CodeFormat = "COMMIT-FORMAT" // not type(scope): subject
	CodeType   = "COMMIT-TYPE"   // type outside the allowed list
	CodeLength = "COMMIT-LENGTH" // subject outside 40 to 50 characters
	CodeCase   = "COMMIT-CASE"   // subject starts with a capital letter
	CodePeriod = "COMMIT-PERIOD" // subject ends with a period
	CodeBody   = "COMMIT-BODY"   // body not separated by a blank line
)

const (
	minLength = 40
	maxLength = 50
)

var types = []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"}

// subjectLine captures the type and what comes after "type(scope): ".
var subjectLine = regexp.MustCompile(`^([a-z]+)(?:\([^()\s]+\))?: (.+)$`)

// generated are the prefixes of messages git writes itself; they are not held to the format.
var generated = []string{"Merge ", `Revert "`, "fixup! ", "squash! ", "amend! "}

// line is a line of the message with its 1-based number in the original text.
type line struct {
	n    int
	text string
}

// Validate checks a message as git hands it to the commit-msg hook: lines starting with # are
// comments and are ignored. name only labels the diagnostics.
//
// shortcut: the imperative mood is not checked, because no rule tells "embed" from "embedded"
// without a dictionary; a reviewer catches it.
func Validate(name, message string) []diag.Diagnostic {
	lines := significant(message)
	if len(lines) == 0 {
		return []diag.Diagnostic{{File: name, Code: CodeEmpty, Message: "the commit message is empty"}}
	}
	subject := lines[0]
	for _, prefix := range generated {
		if strings.HasPrefix(subject.text, prefix) {
			return nil
		}
	}
	out := checkSubject(name, subject)
	if len(lines) > 1 && lines[1].n == subject.n+1 && strings.TrimSpace(lines[1].text) != "" {
		out = append(out, diag.Diagnostic{File: name, Line: lines[1].n, Code: CodeBody, Message: "separate the body from the subject with a blank line"})
	}
	return out
}

// significant returns the lines git keeps: no comments, and nothing before the subject. Blank
// lines inside the message stay, because they separate the subject from the body.
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

func checkSubject(name string, subject line) []diag.Diagnostic {
	var out []diag.Diagnostic
	add := func(code, format string, args ...any) {
		out = append(out, diag.Diagnostic{File: name, Line: subject.n, Code: code, Message: fmt.Sprintf(format, args...)})
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
