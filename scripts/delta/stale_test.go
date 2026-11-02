package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func writeEvent(t *testing.T, root, rel string) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"hook_event_name": "PostToolUse",
		"tool_name":       "Edit",
		"tool_input":      map[string]any{"file_path": filepath.Join(root, rel)},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func commitAll(t *testing.T, p *project, message string) {
	t.Helper()
	if out, err := git(t, p.root, "add", "-A"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if out, err := git(t, p.root, "commit", "-q", "-m", message); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
}

// The whole loop of Phase 5: a slice is verified, its code changes, and delta says so on its own, until the slice
// is built and verified again.
func TestAVerifiedSliceGoesStaleWhenItsCodeChangesAndIsVerifiedAgain(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: git --version")))
	p.write("billing_test.go", "// INV-01\n")
	p.write("main.go", "package main\n")
	commitAll(t, p, "chore(repo): add the billing slice")

	if code, out, _ := p.delta("", "check"); code != 0 || !strings.Contains(out, "billing is verified") {
		t.Fatalf("check = %d %q; want the slice verified", code, out)
	}
	if code, out, errOut := p.delta("", "stale"); code != 0 || out != "" {
		t.Fatalf("stale right after a check = %d %q %q; want nothing to say", code, out, errOut)
	}

	p.write("main.go", "package main\n\nfunc main() {}\n")
	if code, _, errOut := p.delta(writeEvent(t, p.root, "main.go"), "hook"); code != 2 || !strings.Contains(errOut, "billing") || !strings.Contains(errOut, "stale") || !strings.Contains(errOut, "main.go") {
		t.Fatalf("PostToolUse = %d %q; want Claude told that billing went stale because of main.go", code, errOut)
	}
	if code, out, _ := p.delta("", "status"); code != 0 || !strings.Contains(out, "stale (1 files changed)") {
		t.Fatalf("status = %d %q; want billing stale", code, out)
	}
	if code, out, _ := p.delta("", "stale"); code != 1 || !strings.Contains(out, "billing is stale") || !strings.Contains(out, "main.go") {
		t.Fatalf("stale = %d %q; want it to list main.go and fail", code, out)
	}

	// Builder rebuilds it: built, then checked, which is the new baseline.
	if code, _, errOut := p.delta("", "ready", "billing"); code != 0 {
		t.Fatalf("ready = %d %q; a stale slice is ready for Builder", code, errOut)
	}
	if code, _, errOut := p.delta("", "built", "billing"); code != 0 {
		t.Fatalf("built = %d %q", code, errOut)
	}
	if code, out, _ := p.delta("", "check"); code != 0 || !strings.Contains(out, "billing is verified") {
		t.Fatalf("check = %d %q; want the slice verified again", code, out)
	}
	if code, out, errOut := p.delta("", "stale"); code != 0 || out != "" {
		t.Fatalf("stale after the rebuild = %d %q %q; want nothing to say", code, out, errOut)
	}
}

// Nothing was verified on this machine, so there is nothing to compare: a fresh clone is not stale.
func TestASliceThatWasNeverVerifiedHereIsNotStale(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", strings.Replace(built("billing", claimOf("INV-01", "test: git --version")), "status: built", "status: verified", 1))
	p.write("main.go", "package main\n")
	commitAll(t, p, "chore(repo): add the billing slice")
	p.write("main.go", "package main\n\nfunc main() {}\n")
	if code, out, errOut := p.delta("", "stale"); code != 0 || out != "" {
		t.Fatalf("stale = %d %q %q", code, out, errOut)
	}
}

// What delta installs is not code: running update, or editing budget.conf, must not make a slice stale.
func TestChangingDeltasOwnFilesDoesNotMakeASliceStale(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: git --version")))
	p.write("billing_test.go", "// INV-01\n")
	p.write("main.go", "package main\n")
	commitAll(t, p, "chore(repo): add the billing slice")
	if code, out, _ := p.delta("", "check"); code != 0 || !strings.Contains(out, "billing is verified") {
		t.Fatalf("check = %d %q", code, out)
	}
	if code, _, errOut := p.delta("", "init"); code != 0 {
		t.Fatalf("init = %d %q", code, errOut)
	}
	p.write("deltas/budget.conf", "doc_lines=250\n")
	p.write(".claude/settings.json", "{}\n")
	if code, out, errOut := p.delta("", "stale"); code != 0 || out != "" {
		t.Fatalf("stale = %d %q %q; delta's own files are not the slice's code", code, out, errOut)
	}
}
