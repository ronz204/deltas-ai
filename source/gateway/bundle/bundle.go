// Package bundle hands out the files embedded in the harness, and says where each goes in a project.
package bundle

import (
	"bytes"
	"io/fs"
	"path"
	"strings"
	"text/template"

	"github.com/ronz204/deltas-ai/source/harness"
	"github.com/ronz204/deltas-ai/source/usecase/setup"
)

type Bundle struct{}

func (Bundle) Spec() ([]byte, error) { return harness.FS.ReadFile("core/templates/spec.md") }

func (Bundle) GitHook(name string) ([]byte, error) {
	return harness.FS.ReadFile("core/githooks/" + name)
}

// Settings is the part of .claude/settings.json that delta needs: its hooks and permissions.
func (Bundle) Settings() ([]byte, error) { return harness.FS.ReadFile("core/configs/settings.json") }

// Assets lists what init copies into a project, in a fixed order: the skills and the rules under .claude/,
// delta.conf, and a CLAUDE.md that names the project. Rules from any subfolder of rules/ land in .claude/rules/.
func (Bundle) Assets(project string) ([]setup.Asset, error) {
	var out []setup.Asset
	err := fs.WalkDir(harness.FS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		dest, ok := destination(p)
		if !ok {
			return nil
		}
		content, err := harness.FS.ReadFile(p)
		if err != nil {
			return err
		}
		if p == "core/templates/CLAUDE.md" {
			if content, err = fill(content, project); err != nil {
				return err
			}
		}
		out = append(out, setup.Asset{Path: dest, Content: content, Mode: mode(dest)})
		return nil
	})
	return out, err
}

// destination maps a path in the harness to its place in a project; not every file is copied.
func destination(p string) (string, bool) {
	switch {
	case strings.HasPrefix(p, "skills/"):
		return ".claude/" + p, true
	case strings.HasPrefix(p, "rules/"):
		return ".claude/rules/" + path.Base(p), true
	case p == "core/configs/delta.conf":
		return "delta.conf", true
	case p == "core/templates/CLAUDE.md":
		return "CLAUDE.md", true
	}
	return "", false
}

func mode(dest string) fs.FileMode {
	if strings.HasSuffix(dest, ".sh") {
		return 0o755
	}
	return 0o644
}

func fill(text []byte, project string) ([]byte, error) {
	t, err := template.New("CLAUDE.md").Parse(string(text))
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	err = t.Execute(&b, struct{ Project string }{project})
	return b.Bytes(), err
}
