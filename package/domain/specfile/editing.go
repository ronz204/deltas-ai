package specfile

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/claimset"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

func SetStatus(src []byte, to states.Status) ([]byte, error) {
	lines := strings.Split(string(src), "\n")
	end := frontmatterEnd(lines)
	if end < 0 {
		end = len(lines)
	}
	for i := 1; i < end; i++ {
		if m := keyLine.FindStringSubmatch(strings.TrimSuffix(lines[i], "\r")); m != nil && m[1] == "status" {
			return replaceLine(lines, i, "status: "+string(to)), nil
		}
	}
	return nil, errors.New("the frontmatter has no status line; run delta spec lint")
}

func SetTag(src []byte, id claimset.ID, tag claimset.Tag) ([]byte, error) {
	sp, _ := Parse("", src)
	at := slices.IndexFunc(sp.Claims, func(c claimset.Claim) bool { return c.ID == id })
	if at < 0 {
		return nil, fmt.Errorf("claim %s is not among the valid claims of the spec; run delta spec lint", id)
	}
	lines := strings.Split(string(src), "\n")
	return replaceLine(lines, sp.Claims[at].Line-1, "- "+id.String()+" ["+tag.String()+"]"), nil
}

func replaceLine(lines []string, i int, text string) []byte {
	if strings.HasSuffix(lines[i], "\r") {
		text += "\r"
	}
	lines[i] = text
	return []byte(strings.Join(lines, "\n"))
}

func frontmatterEnd(lines []string) int {
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return i
		}
	}
	return -1
}
