package setup

import (
	"context"
	"fmt"
	"github.com/ronz204/deltas-ai/package/domain/conffile"
	"io/fs"
	"strings"
	"testing"
)

var ctx = context.Background()

// world stands in for the project's files and the programs on its PATH.
type world struct {
	files  map[string][]byte
	perms  map[string]fs.FileMode
	writes int
	onPath string
	self   bool
	tools  map[string]bool // other programs on the PATH
	assets []Asset         // what the harness ships; the default two when nil
}

func newWorld() *world {
	return &world{files: map[string][]byte{}, perms: map[string]fs.FileMode{}}
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

func installer(w *world) *Installer { return New(w, w, w, conffile.Conf{}) }

func TestAFreshProjectGetsTheHarnessAndTheSettings(t *testing.T) {
	w := newWorld()
	steps, err := installer(w).Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []Step{{true, "installed 3 files from the harness"}, {true, "recorded the installed harness files in deltas/deltas.manifest"}, {true, "installed .claude/settings.json"}}
	if fmt.Sprint(steps) != fmt.Sprint(want) {
		t.Fatalf("steps = %v, want %v", steps, want)
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

func healthy() *world {
	w := newWorld()
	w.onPath, w.self = "/usr/bin/delta", true
	return w
}

func TestAHealthyInstallationHasNothingToReport(t *testing.T) {
	if got := installer(healthy()).Check(ctx); len(got) != 0 {
		t.Fatalf("Check = %v, want nothing", got)
	}
}

func TestCheckWarnsWhenDeltaIsNotTheOneOnThePath(t *testing.T) {
	w := healthy()
	w.onPath = ""
	if got := installer(w).Check(ctx); len(got) != 1 || got[0] != "delta is not on the PATH, so the hooks of Claude Code cannot run it" {
		t.Fatalf("not on the path: %q", got)
	}
	w.onPath, w.self = "/opt/git-delta/delta", false
	if got := installer(w).Check(ctx); len(got) != 1 || got[0] != "the delta on the PATH (/opt/git-delta/delta) is not this program, so the hooks of Claude Code would run it instead" {
		t.Fatalf("another program: %q", got)
	}
}

func TestCheckReportsWhatTheConfigFilesGotWrong(t *testing.T) {
	w := healthy()
	conf := conffile.Parse("colour=blue\n", "")
	problems := New(w, w, w, conf).Check(ctx)
	if len(problems) != 1 || !strings.Contains(problems[0], `deltas/budget.conf:1: unknown key "colour"`) {
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
