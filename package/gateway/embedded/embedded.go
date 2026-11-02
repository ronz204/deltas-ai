package embedded

import (
	"bytes"
	"io/fs"
	"path"
	"strings"
	"text/template"

	"github.com/ronz204/deltas-ai/agents"
	"github.com/ronz204/deltas-ai/package/usecase/deployer"
)

type Bundle struct{}

func (Bundle) Spec() ([]byte, error) { return agents.FS.ReadFile("core/samples/specify.md") }

func (Bundle) Settings() ([]byte, error) { return agents.FS.ReadFile("core/configs/settings.json") }

func (Bundle) Assets(project string) ([]deployer.Asset, error) {
	var out []deployer.Asset
	err := fs.WalkDir(agents.FS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		dest, ok := destination(p)
		if !ok {
			return nil
		}
		content, err := agents.FS.ReadFile(p)
		if err != nil {
			return err
		}
		if p == "core/samples/CLAUDE.md" || strings.HasPrefix(p, "docs/") {
			if content, err = fill(p, content, project); err != nil {
				return err
			}
		}
		out = append(out, deployer.Asset{Path: dest, Content: content, Managed: managed(p)})
		return nil
	})
	return out, err
}

func managed(p string) bool { return strings.HasPrefix(p, "skills/") || strings.HasPrefix(p, "rules/") }

func destination(p string) (string, bool) {
	switch {
	case strings.HasPrefix(p, "skills/"), strings.HasPrefix(p, "rules/"):
		return ".claude/" + p, true
	case p == "core/configs/budget.conf", p == "core/configs/testing.conf":
		return "deltas/" + path.Base(p), true
	case p == "core/samples/CLAUDE.md":
		return "CLAUDE.md", true
	case strings.HasPrefix(p, "docs/"):
		return "deltas/" + p, true
	}
	return "", false
}

func fill(name string, text []byte, project string) ([]byte, error) {
	t, err := template.New(name).Parse(string(text))
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	err = t.Execute(&b, struct{ Project string }{project})
	return b.Bytes(), err
}
