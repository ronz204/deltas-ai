package bundle

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/ronz204/deltas-ai/agents"
	"github.com/ronz204/deltas-ai/source/domain/lifecycle"
	"github.com/ronz204/deltas-ai/source/domain/specfile"
)

func render(t *testing.T, name string, data any) string {
	t.Helper()
	raw, err := agents.FS.ReadFile("core/templates/" + name)
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

// A CR before the newline would break the hook's shebang on Unix.
func TestCommitMsgHookIsAShellScriptThatCallsDelta(t *testing.T) {
	src, err := Bundle{}.GitHook("commit-msg")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	if !strings.HasPrefix(text, "#!/bin/sh\n") || strings.Contains(text, "\r") || !strings.Contains(text, `exec delta commit-msg "$1"`) {
		t.Fatalf("the hook is not what delta init should install:\n%s", text)
	}
	if !strings.Contains(text, "exit 0") {
		t.Fatal("a missing delta must let the commit through, with a warning")
	}
	if _, err := (Bundle{}).GitHook("nope"); err == nil {
		t.Fatal("an unknown hook was found")
	}
}

// The spec template is what delta spec new will start from, so it must already be a valid spec.
func TestSpecTemplateParses(t *testing.T) {
	src := render(t, "spec.md", struct{ Slice string }{"billing"})
	sp, ds := specfile.Parse("billing.spec.md", []byte(src))
	if len(ds) != 0 {
		t.Fatalf("the template is not a valid spec: %v", ds)
	}
	if sp.Slice != "billing" || sp.Status != lifecycle.Draft || sp.Format != 1 || len(sp.Claims) != 0 {
		t.Fatalf("template read as %+v", sp)
	}
}

// CLAUDE.md is always loaded, so the template stays inside the budget's line ceiling.
func TestClaudeTemplateFitsTheBudget(t *testing.T) {
	src := render(t, "CLAUDE.md", struct{ Project string }{"Acme"})
	if !strings.HasPrefix(src, "# Acme\n") {
		t.Fatalf("the project name was not rendered: %.40q", src)
	}
	if n := strings.Count(src, "\n"); n > 80 {
		t.Fatalf("%d lines; the ceiling for CLAUDE.md is 80", n)
	}
}
