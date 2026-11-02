package deployer

import (
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/package/domain/manifest"
)

const (
	rule  = ".claude/rules/one.md"
	skill = ".claude/skills/x/run.sh"
)

func installed(t *testing.T) *world {
	t.Helper()
	w := newWorld()
	if _, err := installer(w).Run(ctx); err != nil {
		t.Fatal(err)
	}
	return w
}

func ship(w *world, path, content string) {
	assets, _ := w.Assets("p")
	next := make([]Asset, len(assets))
	for i, a := range assets {
		if a.Path == path {
			a.Content = []byte(content)
		}
		next[i] = a
	}
	w.assets = next
}

func steps(t *testing.T, w *world) string {
	t.Helper()
	got, err := installer(w).Update(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, s := range got {
		b.WriteString(s.What + "\n")
	}
	return b.String()
}

func TestInitRecordsWhatItInstalled(t *testing.T) {
	w := installed(t)
	m := manifest.Parse(string(w.files[manifest.Name]))
	for _, path := range []string{rule, skill} {
		if m.Files[path] != manifest.Hash(w.files[path]) {
			t.Errorf("%s is not recorded with its hash: %v", path, m.Files)
		}
	}
	if _, ok := m.Files["CLAUDE.md"]; ok || len(m.Files) != 2 {
		t.Errorf("only the managed files are recorded: %v", m.Files)
	}
}

func TestInitDoesNotRecordAFileTheProjectChanged(t *testing.T) {
	w := newWorld()
	w.files[rule] = []byte("mine\n")
	if _, err := installer(w).Run(ctx); err != nil {
		t.Fatal(err)
	}
	if _, ok := manifest.Parse(string(w.files[manifest.Name])).Files[rule]; ok {
		t.Fatal("a file that differs from the harness was recorded as installed")
	}
}

func TestUpdateOfAnUpToDateProjectChangesNothing(t *testing.T) {
	w := installed(t)
	writes := w.writes
	out := steps(t, w)
	if !strings.Contains(out, "already up to date") || w.writes != writes {
		t.Fatalf("out = %q, writes %d -> %d", out, writes, w.writes)
	}
}

func TestUpdateReplacesAFileTheProjectDidNotTouch(t *testing.T) {
	w := installed(t)
	ship(w, rule, "# new\n")
	if out := steps(t, w); !strings.Contains(out, "updated 1 file") || !strings.Contains(out, rule) {
		t.Fatalf("out = %q", out)
	}
	if string(w.files[rule]) != "# new\n" {
		t.Fatalf("file = %q", w.files[rule])
	}
	if manifest.Parse(string(w.files[manifest.Name])).Files[rule] != manifest.Hash([]byte("# new\n")) {
		t.Fatal("the manifest does not have the new hash")
	}
	if out := steps(t, w); !strings.Contains(out, "already up to date") {
		t.Fatalf("a second update = %q", out)
	}
}

func TestUpdateKeepsWhatTheProjectChanged(t *testing.T) {
	w := installed(t)
	w.files[rule] = []byte("my rule\n")
	ship(w, rule, "# new\n")
	out := steps(t, w)
	if string(w.files[rule]) != "my rule\n" || !strings.Contains(out, "left as they are") || !strings.Contains(out, rule) {
		t.Fatalf("file = %q, out = %q", w.files[rule], out)
	}
	if manifest.Parse(string(w.files[manifest.Name])).Files[rule] == "" {
		t.Fatal("an edited file lost its record, so the next update would treat it as unknown")
	}
}

func TestUpdateKeepsAFileDeltaDoesNotKnowItInstalled(t *testing.T) {
	w := newWorld()
	w.files[rule] = []byte("old\n")
	out := steps(t, w)
	if string(w.files[rule]) != "old\n" || !strings.Contains(out, rule) {
		t.Fatalf("file = %q, out = %q", w.files[rule], out)
	}
}

func TestUpdateInstallsWhatIsMissing(t *testing.T) {
	w := installed(t)
	delete(w.files, skill)
	if out := steps(t, w); !strings.Contains(out, "installed 1 new file") || string(w.files[skill]) != "#!/bin/sh\n" {
		t.Fatalf("out = %q, file = %q", out, w.files[skill])
	}
}

func TestUpdateRemovesWhatTheHarnessNoLongerShipsUnlessTheProjectChangedIt(t *testing.T) {
	w := installed(t)
	w.assets = []Asset{{Path: rule, Content: []byte("# p\n"), Managed: true}}
	out := steps(t, w)
	if _, there := w.files[skill]; there || !strings.Contains(out, "removed 1 file") {
		t.Fatalf("out = %q; the untouched file should be removed", out)
	}
	if _, in := manifest.Parse(string(w.files[manifest.Name])).Files[skill]; in {
		t.Fatal("a removed file is still in the manifest")
	}

	w = installed(t)
	w.files[skill] = []byte("#!/bin/sh\necho mine\n")
	w.assets = []Asset{{Path: rule, Content: []byte("# p\n"), Managed: true}}
	out = steps(t, w)
	if string(w.files[skill]) != "#!/bin/sh\necho mine\n" || !strings.Contains(out, "no longer in the harness, but the project changed them") {
		t.Fatalf("out = %q; an edited file must stay", out)
	}
	if _, in := manifest.Parse(string(w.files[manifest.Name])).Files[skill]; in {
		t.Fatal("a file delta no longer ships is still in the manifest")
	}
}

func TestUpdateNeverTouchesWhatIsInstalledOnce(t *testing.T) {
	w := installed(t)
	w.assets = append(w.assets, Asset{Path: "deltas/budget.conf", Content: []byte("a=1\n")})
	w.files["deltas/budget.conf"] = []byte("doc_lines=250\n")
	steps(t, w)
	if string(w.files["deltas/budget.conf"]) != "doc_lines=250\n" {
		t.Fatalf("the config the project edited was replaced: %q", w.files["deltas/budget.conf"])
	}
}

func TestCheckSaysWhenTheHarnessIsOutOfDate(t *testing.T) {
	w := installed(t)
	if p := installer(w).Check(ctx); hasOutOfDate(p) {
		t.Fatalf("a project that is up to date was flagged: %v", p)
	}
	ship(w, rule, "# new\n")
	if p := installer(w).Check(ctx); !hasOutOfDate(p) {
		t.Fatalf("Check = %v, want the out-of-date warning", p)
	}
	w.files[rule] = []byte("mine\n")
	if p := installer(w).Check(ctx); hasOutOfDate(p) {
		t.Fatalf("an edited file was flagged: %v", p)
	}
}

func TestCheckStaysQuietInAProjectWithNoManifest(t *testing.T) {
	w := newWorld()
	if p := installer(w).Check(ctx); hasOutOfDate(p) {
		t.Fatalf("Check = %v", p)
	}
}

func TestPendingWritesNothing(t *testing.T) {
	w := installed(t)
	ship(w, rule, "# new\n")
	delete(w.files, skill)
	writes := w.writes
	got, err := installer(w).Pending(ctx)
	if err != nil || strings.Join(got, "|") != "install "+skill+"|update "+rule || w.writes != writes {
		t.Fatalf("Pending = %v, %v (writes %d -> %d)", got, err, writes, w.writes)
	}
}

func hasOutOfDate(problems []string) bool {
	for _, p := range problems {
		if strings.Contains(p, "out of date") {
			return true
		}
	}
	return false
}
