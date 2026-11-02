package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func stopEvent(t *testing.T, active bool) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"hook_event_name": "Stop", "stop_hook_active": active})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// The whole loop: Claude changes the tree and tries to finish, delta check runs, and Claude may finish.
func TestStopWaitsForACheckOfTheCurrentTree(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: git --version")))
	p.write("billing_test.go", "// INV-01\n")
	if out, err := git(t, p.root, "add", "-A"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if out, err := git(t, p.root, "commit", "-q", "-m", "chore(repo): add the billing slice"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}

	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 || errOut != "" {
		t.Fatalf("a clean tree: Stop = %d %q, want it allowed", code, errOut)
	}

	p.write("main.go", "package main\n")
	code, _, errOut := p.delta(stopEvent(t, false), "hook")
	if code != 2 || !strings.Contains(errOut, "billing") || !strings.Contains(errOut, "delta check") {
		t.Fatalf("an unchecked change: Stop = %d %q, want a block that asks for delta check", code, errOut)
	}
	if code, _, errOut := p.delta(stopEvent(t, true), "hook"); code != 0 {
		t.Fatalf("Stop while already continuing = %d %q, want it allowed", code, errOut)
	}

	if code, out, errOut := p.delta("", "check"); code != 0 || strings.Contains(errOut, "CHECK-RECORD") || verdict(out, "billing", "INV-01") != "Holds" {
		t.Fatalf("check = %d %q %q", code, out, errOut)
	}
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 || errOut != "" {
		t.Fatalf("after the check: Stop = %d %q, want it allowed", code, errOut)
	}

	p.write("main.go", "package main\n\nfunc main() {}\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 2 || !strings.Contains(errOut, "billing") {
		t.Fatalf("a change after the check: Stop = %d %q, want a block again", code, errOut)
	}
}

func TestStopSaysWhenTheCheckFoundAViolation(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: git no-such-subcommand")))
	p.write("billing_test.go", "// INV-01\n")
	if code, out, _ := p.delta("", "check"); code != 1 || verdict(out, "billing", "INV-01") != "Violated" {
		t.Fatalf("check = %d %q, want the claim violated", code, out)
	}
	code, _, errOut := p.delta(stopEvent(t, false), "hook")
	if code != 2 || !strings.Contains(errOut, "violated") {
		t.Fatalf("Stop = %d %q, want a block about violated claims", code, errOut)
	}
}

// billingProject is a committed repository with one slice whose command names ./billing, and other code beside it.
func billingProject(t *testing.T, conf string) *project {
	t.Helper()
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: "+command("pass", "./billing"))))
	p.write("billing/total_test.go", "// INV-01\n")
	p.write("billing/total.go", "package billing\n")
	p.write("shared/money.go", "package shared\n")
	p.write("payments/pay.go", "package payments\n")
	if conf != "" {
		p.write("deltas/deltas.config", conf)
	}
	if out, err := git(t, p.root, "add", "-A"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if out, err := git(t, p.root, "commit", "-q", "-m", "chore(repo): add the billing slice"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	return p
}

// Without deltas.config a slice depends on every code file, as a command cannot say what its code reaches:
// only documentation is left out.
func TestStopAsksForACheckWhateverCodeChangesButNotDocumentation(t *testing.T) {
	p := billingProject(t, "")

	p.write("README.md", "# a new readme\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 {
		t.Fatalf("a README: Stop = %d %q, want it allowed", code, errOut)
	}
	p.write("shared/money.go", "package shared\n\nfunc Round() {}\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 2 || !strings.Contains(errOut, "billing") {
		t.Fatalf("a shared package: Stop = %d %q, want a block, the command names only ./billing", code, errOut)
	}
}

// deltas.config can say what a slice depends on, and then only that counts.
func TestStopFollowsTheScopeDeltaConfDeclares(t *testing.T) {
	p := billingProject(t, "scope.billing=billing shared\n")

	p.write("payments/pay.go", "package payments\n\nfunc Pay() {}\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 {
		t.Fatalf("code outside the declared paths: Stop = %d %q, want it allowed", code, errOut)
	}
	p.write("shared/money.go", "package shared\n\nfunc Round() {}\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 2 || !strings.Contains(errOut, "billing") {
		t.Fatalf("a declared path: Stop = %d %q, want a block", code, errOut)
	}
	if code, out, _ := p.delta("", "check"); code != 0 || verdict(out, "billing", "INV-01") != "Holds" {
		t.Fatalf("check = %d %q", code, out)
	}
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 {
		t.Fatalf("after the check: Stop = %d %q, want it allowed", code, errOut)
	}
}

func postEdit(t *testing.T, root, rel string) string {
	t.Helper()
	return hookEvent(t, "PostToolUse", "Write", filepath.Join(root, filepath.FromSlash(rel)), "")
}

func TestPostToolUseLintsTheTestFileJustWritten(t *testing.T) {
	p := newProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: "+command("pass", "./billing"))))
	p.write("billing/total_test.go", "// INV-01\nfunc TestTotal(t *testing.T) { t.Skip() }\n")
	p.write("billing/other_test.go", "func TestOther(t *testing.T) { t.Skip() }\n")
	p.write("billing/total.go", "package billing\n")

	code, _, errOut := p.delta(postEdit(t, p.root, "billing/total_test.go"), "hook")
	if code != 2 || !strings.Contains(errOut, "LINT-SKIP") || !strings.Contains(errOut, "billing/total_test.go:2") {
		t.Fatalf("a skipped test that cites a claim: PostToolUse = %d %q, want the problem told to Claude", code, errOut)
	}
	if code, _, errOut := p.delta(postEdit(t, p.root, "billing/total.go"), "hook"); code != 0 || errOut != "" {
		t.Fatalf("a file that is not a test: PostToolUse = %d %q, want it left alone", code, errOut)
	}
	p.write("billing/total_test.go", "// INV-01\nfunc TestTotal(t *testing.T) {}\n")
	if code, _, errOut := p.delta(postEdit(t, p.root, "billing/total_test.go"), "hook"); code != 0 || errOut != "" {
		t.Fatalf("a clean test file: PostToolUse = %d %q, want it left alone", code, errOut)
	}
}

// Forge's turn: code written for an approved slice is pending work until delta built, then delta check closes it.
func TestStopPointsAtForgeUntilTheSliceIsBuilt(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", spec("billing", claimOf("INV-01", "test: git --version")))
	p.write("billing_test.go", "// INV-01\n")
	if out, err := git(t, p.root, "add", "-A"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if out, err := git(t, p.root, "commit", "-q", "-m", "chore(repo): add the billing slice"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 || errOut != "" {
		t.Fatalf("nothing changed: Stop = %d %q, want it allowed", code, errOut)
	}

	p.write("main.go", "package main\n")
	code, _, errOut := p.delta(stopEvent(t, false), "hook")
	if code != 2 || !strings.Contains(errOut, "/forge billing") {
		t.Fatalf("code for an approved slice: Stop = %d %q, want a block that points at /forge", code, errOut)
	}

	if code, _, errOut := p.delta("", "built", "billing"); code != 0 {
		t.Fatalf("built = %d %q", code, errOut)
	}
	code, _, errOut = p.delta(stopEvent(t, false), "hook")
	if code != 2 || !strings.Contains(errOut, "delta check") || strings.Contains(errOut, "/forge") {
		t.Fatalf("a built slice: Stop = %d %q, want a block that asks for delta check", code, errOut)
	}
}
