package bundle

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ronz204/deltas-ai/source/domain/budget"
	"github.com/ronz204/deltas-ai/source/domain/conffile"
	"github.com/ronz204/deltas-ai/source/domain/linkcheck"
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
		".claude/skills/helm/SKILL.md", ".claude/skills/quill/SKILL.md", ".claude/skills/forge/SKILL.md", ".claude/skills/seal/SKILL.md",
		".claude/rules/fidelity.md", ".claude/rules/language.md", ".claude/rules/honesty.md", ".claude/rules/base-scope.md", ".claude/rules/base-reliability.md", "deltas/deltas.config", "CLAUDE.md",
		"deltas/docs/overview.md", "deltas/docs/structure.md", "deltas/docs/expertise.md", "deltas/docs/approach.md",
	} {
		if _, ok := by[want]; !ok {
			t.Errorf("%s is not installed", want)
		}
	}
	for path := range by {
		if strings.HasPrefix(path, ".githooks") || strings.Contains(path, "settings.json") || path == "spec.md" {
			t.Errorf("%s is installed by another route, or not at all", path)
		}
		if strings.Contains(path, "/core/") || strings.HasPrefix(path, "core/") || strings.HasPrefix(path, "docs/") || strings.HasPrefix(path, "rules/") || strings.HasPrefix(path, "skills/") {
			t.Errorf("%s keeps a harness path", path)
		}
	}
}

func TestScriptsAreExecutableAndTheRestIsNot(t *testing.T) {
	for path, a := range assetsOf(t) {
		want := 0o644
		if strings.HasSuffix(path, ".sh") {
			want = 0o755
		}
		if int(a.Mode) != want {
			t.Errorf("%s has mode %o, want %o", path, a.Mode, want)
		}
	}
}

func TestClaudeMdNamesTheProject(t *testing.T) {
	text := string(assetsOf(t)["CLAUDE.md"].Content)
	if !strings.HasPrefix(text, "# acme\n") || strings.Contains(text, "{{") {
		t.Fatalf("CLAUDE.md = %q", text)
	}
}

// The deltas.config that init installs states delta's defaults, so a project starts with what delta would do
// without the file, and the file has to be one delta accepts.
func TestTheDeltaConfThatIsInstalledStatesTheDefaults(t *testing.T) {
	c := conffile.Parse(string(assetsOf(t)["deltas/deltas.config"].Content))
	if len(c.Problems) != 0 {
		t.Fatalf("deltas.config has problems: %v", c.Problems)
	}
	if !reflect.DeepEqual(c.TestFiles, linkcheck.DefaultTestFiles) {
		t.Errorf("test_files = %v, want %v", c.TestFiles, linkcheck.DefaultTestFiles)
	}
	if c.Timeout != 10*time.Minute || len(c.Tools) != 0 || len(c.Scopes) != 0 {
		t.Errorf("timeout = %v, tools = %v, scopes = %v; want 10m and none", c.Timeout, c.Tools, c.Scopes)
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

// The doc templates are the project's once installed: they name the project and delta update leaves them alone.
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
