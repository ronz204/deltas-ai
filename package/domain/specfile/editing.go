package specfile

import (
	"fmt"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/claimset"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

// SetStatus replaces the status in the frontmatter and changes nothing else in the file, line endings included.
func SetStatus(src []byte, to states.Status) ([]byte, error) {
	lines := strings.Split(string(src), "\n")
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			break
		}
		if m := keyLine.FindStringSubmatch(strings.TrimSuffix(lines[i], "\r")); m != nil && m[1] == "status" {
			cr := ""
			if strings.HasSuffix(lines[i], "\r") {
				cr = "\r"
			}
			lines[i] = "status: " + string(to) + cr
			return []byte(strings.Join(lines, "\n")), nil
		}
	}
	return nil, fmt.Errorf("the frontmatter has no status line; run delta spec lint")
}

// SetTag replaces the tag of one claim and changes nothing else in the file, line endings included.
func SetTag(src []byte, id claimset.ID, tag claimset.Tag) ([]byte, error) {
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
