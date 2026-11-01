package specfile

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"

	"github.com/ronz204/deltas-ai/source/domain/findings"
)

var sliceName = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ValidSlice also keeps a name safe to use as a file name.
func ValidSlice(name string) bool { return sliceName.MatchString(name) }

// A slice is named "invoicing", or "billing/invoicing" when its deltas/ folder is not the project's own:
// the part before the last slash is the folder that holds the deltas/ folder.

// SplitRef splits a slice name into that folder (empty for the project's own deltas/) and the bare name.
func SplitRef(ref string) (dir, name string, ok bool) {
	dir, name = "", ref
	if i := strings.LastIndex(ref, "/"); i >= 0 {
		dir, name = ref[:i], ref[i+1:]
		if dir == "" {
			return "", "", false
		}
	}
	if !ValidSlice(name) {
		return "", "", false
	}
	if dir != "" && (dir == "." || !fs.ValidPath(dir) || strings.ContainsAny(dir, `\:`)) {
		return "", "", false
	}
	return dir, name, true
}

// Dir is the folder a slice's commands run from, relative to the project; empty for the project root.
func Dir(ref string) string {
	if i := strings.LastIndex(ref, "/"); i >= 0 {
		return ref[:i]
	}
	return ""
}

// Base is the slice name the frontmatter of its spec carries.
func Base(ref string) string { return ref[strings.LastIndex(ref, "/")+1:] }

// ParseSlice also checks that the frontmatter names the same slice as the file.
func ParseSlice(file, slice string, src []byte) (Spec, []findings.Diagnostic) {
	sp, ds := Parse(file, src)
	if sp.Slice != "" && sp.Slice != Base(slice) {
		ds = append(ds, findings.Diagnostic{File: file, Code: CodeName, Message: fmt.Sprintf("the frontmatter names slice %q but this is the spec of %q", sp.Slice, Base(slice))})
	}
	return sp, ds
}
