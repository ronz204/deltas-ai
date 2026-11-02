package bundle

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/ronz204/deltas-ai/agents"
	"github.com/ronz204/deltas-ai/package/domain/specfile"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

func render(t *testing.T, name string, data any) string {
	t.Helper()
	raw, err := agents.FS.ReadFile("core/samples/" + name)
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := template.New(name).Parse(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, data); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

func TestSpecIsTheSpecTemplate(t *testing.T) {
	src, err := Bundle{}.Spec()
	if err != nil || !strings.Contains(string(src), "format: 1") {
		t.Fatalf("Spec() = %.60q, %v", src, err)
	}
}

func TestSpecTemplateParses(t *testing.T) {
	src := render(t, "specify.md", struct{ Slice string }{"billing"})
	sp, ds := specfile.Parse("billing.spec.md", []byte(src))
	if len(ds) != 0 {
		t.Fatalf("the template is not a valid spec: %v", ds)
	}
	if sp.Slice != "billing" || sp.Status != states.Draft || sp.Format != 1 || len(sp.Claims) != 0 {
		t.Fatalf("template read as %+v", sp)
	}
}

func TestClaudeTemplateFitsTheBudget(t *testing.T) {
	src := render(t, "CLAUDE.md", struct{ Project string }{"Acme"})
	if !strings.HasPrefix(src, "# Acme\n") {
		t.Fatalf("the project name was not rendered: %.40q", src)
	}
	if n := strings.Count(src, "\n"); n > 80 {
		t.Fatalf("%d lines; the ceiling for CLAUDE.md is 80", n)
	}
}
