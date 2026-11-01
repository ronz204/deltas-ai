package budget

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// folded is a description whose text starts on the next lines: description: >-
var folded = regexp.MustCompile(`^[>|][-+]?$`)

// matter is what the weigher needs from a file's frontmatter.
type matter struct {
	end   int  // line number of the closing ---, 0 when there is no frontmatter
	paths bool // it has a paths: key, so it loads only when a matching file is read
	desc  int  // characters of its description
}

// readMatter reads the frontmatter the way the budget script does: the description may be on one
// line or folded over indented lines, and each continuation line counts one extra character.
func readMatter(lines []string) matter {
	var m matter
	inside, inDesc := false, false
	for i, line := range lines {
		n := i + 1
		switch {
		case n == 1 && line == "---":
			inside = true
		case inside && line == "---":
			inside, inDesc, m.end = false, false, n
		case inside && strings.HasPrefix(line, "description:"):
			value := strings.TrimLeft(strings.TrimPrefix(line, "description:"), " \t")
			if folded.MatchString(value) {
				value = ""
			}
			m.desc, inDesc = utf8.RuneCountInString(value), true
		case inside && inDesc && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")):
			m.desc += utf8.RuneCountInString(strings.TrimLeft(line, " \t")) + 1
		case inside:
			inDesc = false
			if strings.HasPrefix(line, "paths:") {
				m.paths = true
			}
		}
	}
	return m
}
