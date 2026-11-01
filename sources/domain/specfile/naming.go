package specfile

import (
	"fmt"
	"regexp"

	"github.com/ronz204/deltas-ai/sources/domain/findings"
)

var sliceName = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ValidSlice also keeps a name safe to use as a file name.
func ValidSlice(name string) bool { return sliceName.MatchString(name) }

// ParseSlice also checks that the frontmatter names the same slice as the file.
func ParseSlice(file, slice string, src []byte) (Spec, []findings.Diagnostic) {
	sp, ds := Parse(file, src)
	if sp.Slice != "" && sp.Slice != slice {
		ds = append(ds, findings.Diagnostic{File: file, Code: CodeName, Message: fmt.Sprintf("the frontmatter names slice %q but this is the spec of %q", sp.Slice, slice)})
	}
	return sp, ds
}
