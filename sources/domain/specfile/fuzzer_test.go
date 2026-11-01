package specfile

import (
	"bytes"
	"testing"

	"github.com/ronz204/deltas-ai/sources/domain/assertion"
)

// FuzzParse must never panic, and what Parse returns must be consistent whatever the input.
func FuzzParse(f *testing.F) {
	f.Add([]byte(validSpec))
	f.Add([]byte(head))
	f.Add([]byte("---\nformat: 1\n"))
	f.Add([]byte("\xef\xbb\xbf---\r\nformat: 1\r\n---\r\n- INV-01 [none]\r\n  When x, y.\r\n"))
	f.Add([]byte(head + "```\n- INV-01 [none]\n<!--\n- ACC-01 [test: x]\n"))
	f.Add([]byte(head + "- INV-01 [test: a[b]]\n\t When x, y.\n- INV-01 [none]\n"))
	f.Fuzz(func(t *testing.T, src []byte) {
		sp, ds := Parse("f.spec.md", src)
		lines := bytes.Count(src, []byte("\n")) + 1
		seen := map[assertion.ID]bool{}
		for _, c := range sp.Claims {
			if c.Line < 1 || c.Line > lines {
				t.Fatalf("claim %s at line %d of %d", c.ID, c.Line, lines)
			}
			if seen[c.ID] {
				t.Fatalf("claim %s returned twice", c.ID)
			}
			seen[c.ID] = true
			if c.Text == "" || c.Tag.Kind == 0 {
				t.Fatalf("claim %s returned half-parsed: %+v", c.ID, c)
			}
		}
		for _, d := range ds {
			if d.Code == "" || d.Line < 0 || d.Line > lines {
				t.Fatalf("bad diagnostic %+v for %d lines", d, lines)
			}
		}
	})
}
