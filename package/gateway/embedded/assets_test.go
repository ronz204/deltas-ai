package embedded

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/package/domain/budget"
	"github.com/ronz204/deltas-ai/package/domain/conffile"
	"github.com/ronz204/deltas-ai/package/domain/linking"
	"github.com/ronz204/deltas-ai/source/usecase/setup"
)

func assetsOf(t *testing.T) map[string]setup.Asset {
	t.Helper()
	list, err := Bundle{}.Assets("acme")
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]setup.Asset{}
	for _, a := range list {
		if _, twice := by[a.Path]; twice {
			t.Errorf("%s is installed twice", a.Path)
		}
		by[a.Path] = a
	}
	return by
}

func TestAssetsGoWhereClaudeCodeLooksForThem(t *testing.T) {
	by := assetsOf(t)
	for _, want := range []string{
		".claude/skills/comper/SKILL.md", ".claude/skills/drafter/SKILL.md", ".claude/skills/builder/SKILL.md", ".claude/skills/auditor/SKILL.md", ".claude/skills/scriber/SKILL.md",
		".claude/rules/claims-fidelity.md", ".claude/rules/commit-messages.md", ".claude/rules/project-language.md", ".claude/rules/done-needs-proof.md", ".claude/rules/code-change-scope.md", ".claude/rules/code-reliability.md",
		"deltas/budget.conf", "deltas/testing.conf", "CLAUDE.md",
		"deltas/docs/overview.md", "deltas/docs/structure.md", "deltas/docs/expertise.md", "deltas/docs/approach.md",
	} {
		if _, ok := by[want]; !ok {
			t.Errorf("%s is not installed", want)
		}
	}
	for path := range by {
		if strings.Contains(path, "settings.json") || path == "specify.md" {
			t.Errorf("%s is installed by another route, or not at all", path)
		}
		if strings.Contains(path, "/core/") || strings.HasPrefix(path, "core/") || strings.HasPrefix(path, "docs/") || strings.HasPrefix(path, "rules/") || strings.HasPrefix(path, "skills/") {
			t.Errorf("%s keeps a harness path", path)
		}
	}
}

func TestClaudeMdNamesTheProject(t *testing.T) {
	text := string(assetsOf(t)["CLAUDE.md"].Content)
	if !strings.HasPrefix(text, "# acme\n") || strings.Contains(text, "{{") {
		t.Fatalf("CLAUDE.md = %q", text)
	}
}

func TestTheConfigFilesThatAreInstalledStateTheDefaults(t *testing.T) {
	by := assetsOf(t)
	c := conffile.Parse(string(by["deltas/budget.conf"].Content), string(by["deltas/testing.conf"].Content))
	if len(c.Problems) != 0 {
		t.Fatalf("the config files have problems: %v", c.Problems)
	}
	if !reflect.DeepEqual(c.Files.Match, linking.DefaultTestFiles) {
		t.Errorf("match = %v, want %v", c.Files.Match, linking.DefaultTestFiles)
	}
	if want := []string{".git", "node_modules", "vendor"}; !reflect.DeepEqual(c.Files.Skip, want) {
		t.Errorf("skip = %v, want %v", c.Files.Skip, want)
	}
	if c.Budget != budget.Defaults() {
		t.Errorf("budget = %+v, want %+v", c.Budget, budget.Defaults())
	}
}

func TestSettingsAreTheFragmentWithTheHooks(t *testing.T) {
	raw, err := Bundle{}.Settings()
	if err != nil || !strings.Contains(string(raw), `"delta hook"`) {
		t.Fatalf("Settings = %q, %v", raw, err)
	}
}

func TestDocTemplatesNameTheProjectAndBelongToIt(t *testing.T) {
	var found int
	for path, a := range assetsOf(t) {
		if !strings.HasPrefix(path, "deltas/docs/") {
			continue
		}
		found++
		text := string(a.Content)
		if a.Managed || strings.Contains(text, "{{") || !strings.Contains(text, "covers: []") || !strings.Contains(text, "reviewed:") || !strings.Contains(text, "## Non-goals") {
			t.Errorf("%s: managed = %v; want a template with covers, reviewed and a Non-goals footer, owned by the project", path, a.Managed)
		}
	}
	if found != 4 {
		t.Errorf("found %d doc templates, want 4", found)
	}
	if !strings.Contains(string(assetsOf(t)["deltas/docs/overview.md"].Content), "of acme.") {
		t.Error("overview.md does not name the project")
	}
}
