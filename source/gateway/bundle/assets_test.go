package bundle

import (
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/conffile"
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
		".claude/skills/quill/SKILL.md", ".claude/skills/forge/SKILL.md", ".claude/skills/seal/SKILL.md",
		".claude/skills/archivist/SKILL.md", ".claude/skills/surveyor/SKILL.md",
		".claude/skills/archivist/references/culture.md",
		".claude/rules/commit-messages.md", "delta.conf", "CLAUDE.md",
	} {
		if _, ok := by[want]; !ok {
			t.Errorf("%s is not installed", want)
		}
	}
	for path := range by {
		if strings.HasPrefix(path, ".githooks") || strings.Contains(path, "settings.json") || path == "spec.md" {
			t.Errorf("%s is installed by another route, or not at all", path)
		}
		if strings.Contains(path, "/core/") || strings.HasPrefix(path, "core/") || strings.HasPrefix(path, "rules/") || strings.HasPrefix(path, "skills/") {
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

// The file init installs is read by delta itself, so it has to be one delta accepts.
func TestTheDeltaConfThatIsInstalledHasNoProblems(t *testing.T) {
	c := conffile.Parse(string(assetsOf(t)["delta.conf"].Content))
	if len(c.Problems) != 0 || c.Budget.Total != 3500 || c.Budget.Rule != 50 {
		t.Fatalf("delta.conf: %+v", c)
	}
}

func TestSettingsAreTheFragmentWithTheHooks(t *testing.T) {
	raw, err := Bundle{}.Settings()
	if err != nil || !strings.Contains(string(raw), `"delta hook"`) {
		t.Fatalf("Settings = %q, %v", raw, err)
	}
}
