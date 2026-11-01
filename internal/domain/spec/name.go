package spec

import (
	"fmt"
	"regexp"

	"github.com/ronz204/deltas-ai/internal/domain/diag"
)

var sliceName = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ValidSlice reports whether name is a slice name: lowercase words joined by hyphens. It is also
// what keeps a name safe to use as a file name.
func ValidSlice(name string) bool { return sliceName.MatchString(name) }

// ParseSlice parses the spec stored for a slice and also checks that its frontmatter names that
// same slice, so a file cannot quietly describe another capability.
func ParseSlice(file, slice string, src []byte) (Spec, []diag.Diagnostic) {
	sp, ds := Parse(file, src)
	if sp.Slice != "" && sp.Slice != slice {
		ds = append(ds, diag.Diagnostic{File: file, Code: CodeName, Message: fmt.Sprintf("the frontmatter names slice %q but this is the spec of %q", sp.Slice, slice)})
	}
	return sp, ds
}
