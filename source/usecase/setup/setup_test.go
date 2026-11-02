package setup

import (
	"context"
	"errors"
	"fmt"
	"github.com/ronz204/deltas-ai/source/domain/conffile"
	"io/fs"
	"strings"
	"testing"
)

var ctx = context.Background()

// world stands in for the project's files, its git configuration and the embedded hooks.
type world struct {
	files    map[string][]byte
	perms    map[string]fs.FileMode
	hooksDir string
	notARepo bool
	writes   int
	onPath   string
	self     bool
	tools    map[string]bool // other programs on the PATH
	assets   []Asset         // what the harness ships; the default two when nil
}

func newWorld() *world {
	return &world{files: map[string][]byte{}, perms: map[string]fs.FileMode{}}
}

const hook = "#!/bin/sh\nexec delta commit-msg \"$1\"\n"

func (w *world) GitHook(name string) ([]byte, error) {
	if name != "commit-msg" {
		return nil, errors.New("unknown hook")
	}
	return []byte(hook), nil
}

func (w *world) ReadFile(_ context.Context, rel string) ([]byte, error) {
	if b, ok := w.files[rel]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("read %s: %w", rel, fs.ErrNotExist)
}

func (w *world) WriteFile(_ context.Context, rel string, content []byte, perm fs.FileMode) error {
	w.files[rel], w.perms[rel] = content, perm
	w.writes++
	return nil
}

func (w *world) RemoveFile(_ context.Context, rel string) error {
	delete(w.files, rel)
	delete(w.perms, rel)
	w.writes++
	return nil
}

func (w *world) Locate(name string) (string, bool) {
	switch {
	case name == "delta":
		return w.onPath, w.self
	case w.tools[name]:
		return "/usr/bin/" + name, false
	}
	return "", false
}

func (w *world) HooksPath(context.Context) (string, error) { return w.hooksDir, nil }

func (w *world) SetHooksPath(_ context.Context, path string) error {
	if w.notARepo {
		return errors.New("git config: not a git repository")
	}
	w.hooksDir = path
	return nil
}

func installer(w *world) *Installer { return New(w, w, w, w, conffile.Conf{}) }

func TestAFreshProjectGetsTheHookAndThePath(t *testing.T) {
	w := newWorld()
	steps, err := installer(w).Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []Step{{true, "installed .githooks/commit-msg"}, {true, "set core.hooksPath to .githooks"}, {true, "installed 3 files from the harness"}, {true, "recorded the installed harness files in deltas/deltas.manifest"}, {true, "installed .claude/settings.json"}}
	if fmt.Sprint(steps) != fmt.Sprint(want) {
		t.Fatalf("steps = %v, want %v", steps, want)
	}
	if string(w.files[".githooks/commit-msg"]) != hook || w.perms[".githooks/commit-msg"] != 0o755 || w.hooksDir != ".githooks" {
		t.Fatalf("hook = %q, mode = %v, hooksPath = %q", w.files[".githooks/commit-msg"], w.perms[".githooks/commit-msg"], w.hooksDir)
	}
}

func TestAFreshProjectGetsTheFilesOfTheHarness(t *testing.T) {
	w := newWorld()
	if _, err := installer(w).Run(ctx); err != nil {
		t.Fatal(err)
	}
	if string(w.files[".claude/rules/one.md"]) != "# acme\n" {
		t.Errorf("rule = %q, want the project's name in it", w.files[".claude/rules/one.md"])
	}
	if w.perms[".claude/skills/x/run.sh"] != 0o755 {
		t.Errorf("a script lost its execute bit: %v", w.perms[".claude/skills/x/run.sh"])
	}
	if _, ok := w.files["deltas/specs/.gitkeep"]; !ok {
		t.Error("the folder for the contracts was not created")
	}
	if string(w.files[".claude/settings.json"]) != fragment {
		t.Errorf("settings = %q", w.files[".claude/settings.json"])
	}
}

func TestAHarnessFileTheProjectChangedIsKeptAndSaid(t *testing.T) {
	w := newWorld()
	w.files[".claude/rules/one.md"] = []byte("# mine\n")
	steps, err := installer(w).Run(ctx)
	if err != nil || string(w.files[".claude/rules/one.md"]) != "# mine\n" {
		t.Fatalf("rule = %q, err = %v", w.files[".claude/rules/one.md"], err)
	}
	var said bool
	for _, s := range steps {
		said = said || (!s.Done && strings.Contains(s.What, ".claude/rules/one.md") && strings.Contains(s.What, "left as they are"))
	}
	if !said {
		t.Fatalf("steps = %v, want the kept file named", steps)
	}
}

func TestSettingsOfTheProjectGainDeltasHooks(t *testing.T) {
	w := newWorld()
	w.files[".claude/settings.json"] = []byte(`{"model":"opus"}`)
	steps, err := installer(w).Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got := decode(t, w.files[".claude/settings.json"])
	if got["model"] != "opus" || got["hooks"] == nil || got["permissions"] == nil {
		t.Fatalf("settings = %s", w.files[".claude/settings.json"])
	}
	if last := steps[len(steps)-1]; !last.Done || !strings.Contains(last.What, "added delta's hooks") {
		t.Fatalf("step = %+v", last)
	}
}

func TestRunningAgainChangesNothing(t *testing.T) {
	w := newWorld()
	if _, err := installer(w).Run(ctx); err != nil {
		t.Fatal(err)
	}
	writes := w.writes
	steps, err := installer(w).Run(ctx)
	if err != nil || w.writes != writes {
		t.Fatalf("a second run wrote again: %v, %d writes", err, w.writes-writes)
	}
	for _, s := range steps {
		if s.Done {
			t.Errorf("a second run reported work done: %+v", s)
		}
	}
}

func TestAnotherHooksPathIsNeverOverridden(t *testing.T) {
	for _, existing := range []string{".husky", "/company/hooks", "tools/git-hooks"} {
		w := newWorld()
		w.hooksDir = existing
		steps, err := installer(w).Run(ctx)
		if err != nil || w.hooksDir != existing {
			t.Fatalf("hooksPath %q became %q (%v)", existing, w.hooksDir, err)
		}
		if steps[1].Done || steps[1].What == "" {
			t.Errorf("%q: the path step = %+v, want it left alone with an explanation", existing, steps[1])
		}
	}
}

func TestADotSlashSpellingOfTheSamePathCounts(t *testing.T) {
	w := newWorld()
	w.hooksDir = "./.githooks"
	if steps, err := installer(w).Run(ctx); err != nil || steps[1].Done || w.hooksDir != "./.githooks" {
		t.Fatalf("steps = %v, hooksPath = %q, err = %v", steps, w.hooksDir, err)
	}
}

func TestADifferentHookIsKept(t *testing.T) {
	w := newWorld()
	w.files[".githooks/commit-msg"] = []byte("#!/bin/sh\necho mine\n")
	steps, err := installer(w).Run(ctx)
	if err != nil || steps[0].Done || string(w.files[".githooks/commit-msg"]) != "#!/bin/sh\necho mine\n" {
		t.Fatalf("steps = %v, hook = %q, err = %v", steps, w.files[".githooks/commit-msg"], err)
	}
}

func TestNothingIsWrittenOutsideAGitRepository(t *testing.T) {
	w := newWorld()
	w.notARepo = true
	if _, err := installer(w).Run(ctx); err == nil {
		t.Fatal("init succeeded outside a repository")
	}
	if w.writes != 0 || len(w.files) != 0 {
		t.Fatalf("files were left behind: %v", w.files)
	}
}

func healthy() *world {
	w := newWorld()
	w.files[".githooks/commit-msg"] = []byte(hook)
	w.hooksDir = ".githooks"
	w.onPath, w.self = "/usr/bin/delta", true
	return w
}

func TestAHealthyInstallationHasNothingToReport(t *testing.T) {
	if got := installer(healthy()).Check(ctx); len(got) != 0 {
		t.Fatalf("Check = %v, want nothing", got)
	}
}

func TestCheckExplainsEachWayTheHookCanBeMissing(t *testing.T) {
	cases := []struct {
		name  string
		tweak func(*world)
		want  string
	}{
		{"nothing installed", func(w *world) { w.hooksDir = ""; delete(w.files, ".githooks/commit-msg") }, "git does not check commit messages yet: run delta init"},
		{"hook present but git not pointed at it", func(w *world) { w.hooksDir = "" }, "core.hooksPath is not set, so git ignores .githooks/commit-msg: run delta init"},
		{"pointed but the hook is gone", func(w *world) { delete(w.files, ".githooks/commit-msg") }, ".githooks/commit-msg is missing: run delta init"},
		{"git points somewhere else", func(w *world) { w.hooksDir = ".husky" }, "core.hooksPath is .husky, so .githooks/commit-msg does not run on its own"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := healthy()
			c.tweak(w)
			got := installer(w).Check(ctx)
			if len(got) != 1 || got[0] != c.want {
				t.Fatalf("Check = %q, want [%q]", got, c.want)
			}
		})
	}
}

func TestCheckWarnsWhenDeltaIsNotTheOneOnThePath(t *testing.T) {
	w := healthy()
	w.onPath = ""
	if got := installer(w).Check(ctx); len(got) != 1 || got[0] != "delta is not on the PATH, so the git hook cannot run it" {
		t.Fatalf("not on the path: %q", got)
	}
	w.onPath, w.self = "/opt/git-delta/delta", false
	if got := installer(w).Check(ctx); len(got) != 1 || got[0] != "the delta on the PATH (/opt/git-delta/delta) is not this program, so the git hook would run it instead" {
		t.Fatalf("another program: %q", got)
	}
}

func TestCheckReportsWhatItCannotReadInsteadOfFailing(t *testing.T) {
	w := healthy()
	got := New(w, brokenGit{}, w, w, conffile.Conf{}).Check(ctx)
	if len(got) != 1 || got[0] != "could not read git's configuration: no git" {
		t.Fatalf("Check = %q", got)
	}
}

type brokenGit struct{}

func (brokenGit) HooksPath(context.Context) (string, error) {
	return "", errors.New("no git")
}
func (brokenGit) SetHooksPath(context.Context, string) error {
	return errors.New("no git")
}

func TestCheckListsTheToolsDeltaConfRequires(t *testing.T) {
	w := healthy()
	conf := conffile.Parse("required_tools=git, docker\n")
	w.tools = map[string]bool{"git": true}
	problems := New(w, w, w, w, conf).Check(ctx)
	if len(problems) != 1 || !strings.Contains(problems[0], "docker is required by deltas.config") {
		t.Fatalf("problems = %q, want only docker", problems)
	}
}

func TestCheckReportsWhatDeltaConfGotWrong(t *testing.T) {
	w := healthy()
	conf := conffile.Parse("colour=blue\n")
	problems := New(w, w, w, w, conf).Check(ctx)
	if len(problems) != 1 || !strings.Contains(problems[0], `deltas/deltas.config:1: unknown key "colour"`) {
		t.Fatalf("problems = %q", problems)
	}
}

func (w *world) Name() string { return "acme" }

func (w *world) Assets(project string) ([]Asset, error) {
	if w.assets != nil {
		return w.assets, nil
	}
	return []Asset{
		{Path: ".claude/rules/one.md", Content: []byte("# " + project + "\n"), Mode: 0o644, Managed: true},
		{Path: ".claude/skills/x/run.sh", Content: []byte("#!/bin/sh\n"), Mode: 0o755, Managed: true},
	}, nil
}

const fragment = `{"permissions":{"deny":["Read(**/.env)"]},"hooks":{"Stop":[{"hooks":[{"type":"command","command":"delta hook","timeout":20}]}]}}`

func (w *world) Settings() ([]byte, error) { return []byte(fragment), nil }
