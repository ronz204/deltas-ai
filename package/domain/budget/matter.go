package budget

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var folded = regexp.MustCompile(`^[>|][-+]?$`)

type matter struct {
	end   int
	paths bool
	desc  int
}

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
