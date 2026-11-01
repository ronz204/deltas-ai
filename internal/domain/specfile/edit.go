package specfile

import (
	"fmt"
	"strings"

	"github.com/ronz204/deltas-ai/internal/domain/assertion"
)

// SetTag returns src with the tag of one claim replaced. Only that header line changes, so the
// rest of the file, its line endings included, comes back as it was.
func SetTag(src []byte, id assertion.ID, tag assertion.Tag) ([]byte, error) {
	sp, _ := Parse("", src)
	line := 0
	for _, c := range sp.Claims {
		if c.ID == id {
			line = c.Line
		}
	}
	if line == 0 {
		return nil, fmt.Errorf("claim %s is not among the valid claims of the spec; run delta spec lint", id)
	}
	lines := strings.Split(string(src), "\n")
	cr := ""
	if strings.HasSuffix(lines[line-1], "\r") {
		cr = "\r"
	}
	lines[line-1] = "- " + id.String() + " [" + tag.String() + "]" + cr
	return []byte(strings.Join(lines, "\n")), nil
}
