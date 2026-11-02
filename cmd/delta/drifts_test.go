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

func TestAVerifiedSliceShowsDriftWhenItsCodeChangesAndACheckClearsIt(t *testing.T) {
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
	verified := p.read("deltas/specs/billing.spec.md")

	p.write("main.go", "package main\n\nfunc main() {}\n")
	if code, out, errOut := p.delta(writeEvent(t, p.root, "main.go"), "hook"); code != 0 || out != "" || errOut != "" {
		t.Fatalf("PostToolUse = %d %q %q; editing code is ordinary work, so Claude hears nothing", code, out, errOut)
	}
	if code, out, _ := p.delta("", "status"); code != 0 || !strings.Contains(out, "verified (1 files changed)") {
		t.Fatalf("status = %d %q; want billing verified with its drift", code, out)
	}
	if code, out, _ := p.delta("", "stale"); code != 1 || !strings.Contains(out, "billing: 1 files changed since it was verified") || !strings.Contains(out, "main.go") {
		t.Fatalf("stale = %d %q; want it to list main.go and fail", code, out)
	}
	if got := p.read("deltas/specs/billing.spec.md"); got != verified {
		t.Fatalf("the spec changed while only the code did:\n%s", got)
	}

	if code, out, _ := p.delta("", "check"); code != 0 || verdict(out, "billing", "INV-01") != "Holds" {
		t.Fatalf("check = %d %q; want the claim to hold", code, out)
	}
	if code, out, errOut := p.delta("", "stale"); code != 0 || out != "" {
		t.Fatalf("stale after the check = %d %q %q; want nothing to say", code, out, errOut)
	}
	if got := p.read("deltas/specs/billing.spec.md"); got != verified {
		t.Fatalf("the spec changed by a check that passed:\n%s", got)
	}
}

func TestAVerifiedSliceCanBeRebuiltWithBuilder(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: git --version")))
	p.write("billing_test.go", "// INV-01\n")
	commitAll(t, p, "chore(repo): add the billing slice")
	if code, out, _ := p.delta("", "check"); code != 0 || !strings.Contains(out, "billing is verified") {
		t.Fatalf("check = %d %q", code, out)
	}
	if code, _, errOut := p.delta("", "ready", "billing"); code != 0 {
		t.Fatalf("ready = %d %q; a verified slice is ready for Builder", code, errOut)
	}
	if code, _, errOut := p.delta("", "built", "billing"); code != 0 {
		t.Fatalf("built = %d %q; a verified slice can be rebuilt", code, errOut)
	}
	if code, out, _ := p.delta("", "check"); code != 0 || !strings.Contains(out, "billing is verified") {
		t.Fatalf("check = %d %q; want the slice verified again", code, out)
	}
}

func TestAViolatedClaimLeavesTheSpecOfAVerifiedSliceAlone(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: git --version")))
	p.write("billing_test.go", "// INV-01\n")
	commitAll(t, p, "chore(repo): add the billing slice")
	if code, _, _ := p.delta("", "check"); code != 0 {
		t.Fatal("check failed")
	}
	verified := p.read("deltas/specs/billing.spec.md")

	p.write("deltas/specs/billing.spec.md", strings.Replace(verified, "test: git --version", "test: git no-such-subcommand", 1))
	if code, out, _ := p.delta("", "check"); code != 1 || verdict(out, "billing", "INV-01") != "Violated" {
		t.Fatalf("check = %d %q; want the claim violated", code, out)
	}
	if got := p.read("deltas/specs/billing.spec.md"); !strings.Contains(got, "status: verified") {
		t.Fatalf("a failing check rewrote the status:\n%s", got)
	}
}

func TestASliceMarkedStaleByAnOlderDeltaGoesBackToVerifiedWhenItsTestsPass(t *testing.T) {
	p := gitProject(t)
	legacy := strings.Replace(built("billing", claimOf("INV-01", "test: git --version")), "status: built", "status: stale", 1)
	p.write("deltas/specs/billing.spec.md", legacy)
	p.write("billing_test.go", "// INV-01\n")
	commitAll(t, p, "chore(repo): add the billing slice")
	if code, out, _ := p.delta("", "check"); code != 0 || !strings.Contains(out, "billing is verified") {
		t.Fatalf("check = %d %q; want the stale slice verified", code, out)
	}
	if got := p.read("deltas/specs/billing.spec.md"); !strings.Contains(got, "status: verified") {
		t.Fatalf("the spec is still stale:\n%s", got)
	}
}

func TestAChangeOutsideWhatTheCommandsNameDoesNotMakeASliceStale(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: "+command("pass", "./billing"))))
	p.write("billing/billing_test.go", "// INV-01\n")
	p.write("other/main.go", "package other\n")
	commitAll(t, p, "chore(repo): add the billing slice")
	if code, _, _ := p.delta("", "check"); code != 0 {
		t.Fatal("check failed")
	}

	p.write("other/main.go", "package other\n\nfunc F() {}\n")
	if code, out, errOut := p.delta("", "stale"); code != 0 || out != "" {
		t.Fatalf("stale = %d %q %q; other/ is not part of billing", code, out, errOut)
	}
	p.write("billing/total.go", "package billing\n")
	if code, out, _ := p.delta("", "stale"); code != 1 || !strings.Contains(out, "billing/total.go") {
		t.Fatalf("stale = %d %q; a new file in billing/ is drift", code, out)
	}
}

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
