package linking

import (
	"path"
	"regexp"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/claimset"
)

type File struct{ Path, Text string }

var scriptExts = map[string]bool{
	".go": true, ".py": true, ".js": true, ".jsx": true, ".ts": true, ".tsx": true,
	".mjs": true, ".cjs": true, ".mts": true, ".cts": true,
}

func idPattern(id claimset.ID) *regexp.Regexp {
	prefix, num, _ := strings.Cut(id.String(), "-")
	return regexp.MustCompile(`(?i)` + prefix + `[-_]?` + num + `(?:[^0-9]|$)`)
}

func Cited(id claimset.ID, command string, files []File) bool {
	paths := pathsIn(command)
	re := idPattern(id)
	for _, f := range files {
		if underAny(f.Path, paths) && re.MatchString(f.Text) {
			return true
		}
	}
	return false
}

func pathsIn(command string) []string {
	var paths []string
	for _, arg := range strings.Fields(command) {
		if p, ok := asPath(arg); ok {
			paths = append(paths, p)
		}
	}
	return paths
}

func asPath(arg string) (string, bool) {
	if strings.HasPrefix(arg, "-") {
		return "", false
	}
	p := strings.ReplaceAll(arg, `\`, "/")
	startsHere := p == "." || strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../")
	namesScript := strings.Contains(p, "/") && scriptExts[path.Ext(p)]
	if !startsHere && !namesScript {
		return "", false
	}
	p = strings.TrimSuffix(p, "...")
	p = strings.TrimSuffix(p, "/")
	return path.Clean(p), true
}

func underAny(file string, paths []string) bool {
	if len(paths) == 0 {
		return true
	}
	for _, p := range paths {
		if p == "." || file == p || strings.HasPrefix(file, p+"/") {
			return true
		}
	}
	return false
}
