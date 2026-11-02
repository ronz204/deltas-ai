package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sessionEvent(t *testing.T, name string, active bool) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"hook_event_name": name, "session_id": "s1", "stop_hook_active": active})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func stopEvent(t *testing.T, active bool) string {
	t.Helper()
	return sessionEvent(t, "Stop", active)
}

func startSession(t *testing.T, p *project) {
	t.Helper()
	if code, _, errOut := p.delta(sessionEvent(t, "SessionStart", false), "hook"); code != 0 {
		t.Fatalf("SessionStart = %d %q", code, errOut)
	}
}

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

	startSession(t, p)
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 || errOut != "" {
		t.Fatalf("nothing changed: Stop = %d %q, want it allowed", code, errOut)
	}

	p.write("main.go", "package main\n")
	code, _, errOut := p.delta(stopEvent(t, false), "hook")
	if code != 2 || !strings.Contains(errOut, "delta check billing") {
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
	startSession(t, p)
	p.write("main.go", "package main\n")
	if code, out, _ := p.delta("", "check"); code != 1 || verdict(out, "billing", "INV-01") != "Violated" {
		t.Fatalf("check = %d %q, want the claim violated", code, out)
	}
	code, _, errOut := p.delta(stopEvent(t, false), "hook")
	if code != 2 || !strings.Contains(errOut, "violated") {
		t.Fatalf("Stop = %d %q, want a block about violated claims", code, errOut)
	}
}

func billingProject(t *testing.T) *project {
	t.Helper()
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: "+command("pass", "./billing"))))
	p.write("billing/total_test.go", "// INV-01\n")
	p.write("billing/total.go", "package billing\n")
	p.write("shared/money.go", "package shared\n")
	p.write("payments/pay.go", "package payments\n")
	if out, err := git(t, p.root, "add", "-A"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if out, err := git(t, p.root, "commit", "-q", "-m", "chore(repo): add the billing slice"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	return p
}

func TestStopAsksForACheckOnlyWhenTheCodeOfTheSliceChanges(t *testing.T) {
	p := billingProject(t)
	startSession(t, p)

	p.write("README.md", "# a new readme\n")
	p.write("shared/money.go", "package shared\n\nfunc Round() {}\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 {
		t.Fatalf("a README and code the command does not name: Stop = %d %q, want it allowed", code, errOut)
	}
	p.write("billing/total.go", "package billing\n\nfunc Total() {}\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 2 || !strings.Contains(errOut, "billing") {
		t.Fatalf("code under ./billing: Stop = %d %q, want a block", code, errOut)
	}
	if code, out, _ := p.delta("", "check"); code != 0 || verdict(out, "billing", "INV-01") != "Holds" {
		t.Fatalf("check = %d %q", code, out)
	}
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 {
		t.Fatalf("after the check: Stop = %d %q, want it allowed", code, errOut)
	}
}

func TestStopAsksForACheckWhateverCodeChangesWhenTheCommandNamesNoPath(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: git --version")))
	p.write("billing_test.go", "// INV-01\n")
	p.write("shared/money.go", "package shared\n")
	commitAll(t, p, "chore(repo): add the billing slice")
	startSession(t, p)
	p.write("shared/money.go", "package shared\n\nfunc Round() {}\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 2 || !strings.Contains(errOut, "billing") {
		t.Fatalf("a command with no path: Stop = %d %q, want a block for any code", code, errOut)
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

func TestStopLeavesApprovedSlicesAloneWhateverIsEdited(t *testing.T) {
	p := gitProject(t)
	for _, slice := range []string{"job-claiming", "job-lifecycle", "retry-policy", "worker-runtime"} {
		p.write("deltas/specs/"+slice+".spec.md", spec(slice, claimOf("INV-01", "none")))
	}
	commitAll(t, p, "chore(repo): add the slices")
	startSession(t, p)

	p.write("go.mod", "module jobs\n")
	p.write("image.png", "not really a png")
	p.write("internal/worker/runtime.go", "package worker\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 || errOut != "" {
		t.Fatalf("approved slices and edits: Stop = %d %q, want it allowed: an approved slice is backlog", code, errOut)
	}
}

func TestStopOnlyLooksAtWhatThisSessionChanged(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: git --version")))
	p.write("billing_test.go", "// INV-01\n")
	p.write("main.go", "package main\n")
	commitAll(t, p, "chore(repo): add the billing slice")
	p.write("main.go", "package main\n\nfunc main() {}\n")

	startSession(t, p)
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 || errOut != "" {
		t.Fatalf("code that was already edited when the session began: Stop = %d %q, want it allowed", code, errOut)
	}
	p.write("main.go", "package main\n\nfunc main() { println() }\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 2 || !strings.Contains(errOut, "delta check billing") {
		t.Fatalf("the same file edited again: Stop = %d %q, want a block", code, errOut)
	}
}

func TestStopWithoutASessionStartLetsClaudeFinish(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: git --version")))
	p.write("billing_test.go", "// INV-01\n")
	commitAll(t, p, "chore(repo): add the billing slice")
	p.write("main.go", "package main\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 || errOut != "" {
		t.Fatalf("no baseline: Stop = %d %q, want it allowed", code, errOut)
	}
}

func TestCheckTakesSeveralSlices(t *testing.T) {
	p := gitProject(t)
	for _, slice := range []string{"billing", "payments", "refunds"} {
		p.write("deltas/specs/"+slice+".spec.md", built(slice, claimOf("INV-01", "test: git --version")))
		p.write(slice+"_test.go", "// INV-01\n")
	}
	code, out, errOut := p.delta("", "check", "billing", "payments")
	if code != 0 || verdict(out, "billing", "INV-01") != "Holds" || verdict(out, "payments", "INV-01") != "Holds" || verdict(out, "refunds", "INV-01") != "" {
		t.Fatalf("check billing payments = %d %q %q; want those two and not refunds", code, out, errOut)
	}
}

func TestAFootprintFollowsASliceWhoseTestsLiveApartFromItsCode(t *testing.T) {
	p := gitProject(t)
	p.write("go.mod", "module jobs\n")
	p.write("src/billing/total.go", "package billing\n")
	p.write("src/payments/pay.go", "package payments\n")
	p.write("deltas/specs/billing.spec.md", spec("billing", claimOf("INV-01", "test: git --version")))
	commitAll(t, p, "chore(repo): add the billing slice")

	if code, _, errOut := p.delta("", "ready", "billing"); code != 0 {
		t.Fatalf("ready = %d %q", code, errOut)
	}
	p.write("src/billing/total.go", "package billing\n\nfunc Total() {}\n")
	p.write("testing/unit/billing_test.go", "// INV-01\n")
	code, out, errOut := p.delta("", "built", "billing")
	if code != 0 || !strings.Contains(out, "billing footprint: 2 files") || errOut != "" {
		t.Fatalf("built = %d %q %q", code, out, errOut)
	}
	if got := p.read("deltas/specs/billing.files"); got != "src/billing/total.go\ntesting/unit/billing_test.go\n" {
		t.Fatalf("footprint = %q", got)
	}
	if code, out, _ := p.delta("", "check"); code != 0 || !strings.Contains(out, "billing is verified") {
		t.Fatalf("check = %d %q", code, out)
	}
	commitAll(t, p, "feat(billing): build the billing slice")
	startSession(t, p)

	p.write("go.mod", "module jobs\n\ngo 1.24\n")
	p.write("src/payments/pay.go", "package payments\n\nfunc Pay() {}\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 || errOut != "" {
		t.Fatalf("go.mod and another slice's code: Stop = %d %q, want it allowed", code, errOut)
	}
	if code, out, _ := p.delta("", "stale"); code != 0 || out != "" {
		t.Fatalf("stale = %d %q; only the footprint of billing is its code", code, out)
	}

	p.write("src/billing/total.go", "package billing\n\nfunc Total() int { return 1 }\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 2 || !strings.Contains(errOut, "delta check billing") {
		t.Fatalf("code of billing: Stop = %d %q, want a block that asks for the check", code, errOut)
	}
	if code, out, _ := p.delta("", "stale"); code != 1 || !strings.Contains(out, "src/billing/total.go") {
		t.Fatalf("stale = %d %q; want the file of the footprint listed", code, out)
	}
}

func TestRebuildingASliceGrowsItsFootprint(t *testing.T) {
	p := gitProject(t)
	p.write("src/billing/total.go", "package billing\n")
	p.write("deltas/specs/billing.spec.md", spec("billing", claimOf("INV-01", "none")))
	commitAll(t, p, "chore(repo): add the billing slice")

	build := func(files ...string) {
		t.Helper()
		if code, _, errOut := p.delta("", "ready", "billing"); code != 0 {
			t.Fatalf("ready = %d %q", code, errOut)
		}
		for _, f := range files {
			p.write(f, "package billing\n// "+f+"\n")
		}
		if code, _, errOut := p.delta("", "built", "billing"); code != 0 {
			t.Fatalf("built = %d %q", code, errOut)
		}
	}
	build("src/billing/total.go", "src/billing/tax.go")
	build("src/billing/refund.go")
	if got := p.read("deltas/specs/billing.files"); got != "src/billing/refund.go\nsrc/billing/tax.go\nsrc/billing/total.go\n" {
		t.Fatalf("footprint = %q, want the files of both builds", got)
	}
}

func TestBuiltWithoutReadyKeepsTheFootprintAndSaysSo(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", spec("billing", claimOf("INV-01", "none")))
	commitAll(t, p, "chore(repo): add the billing slice")
	code, out, errOut := p.delta("", "built", "billing")
	if code != 0 || !strings.Contains(out, "billing is built") || !strings.Contains(errOut, "no snapshot from delta ready billing") {
		t.Fatalf("built = %d %q %q", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(p.root, "deltas", "specs", "billing.files")); !os.IsNotExist(err) {
		t.Fatalf("a footprint was written without a snapshot: %v", err)
	}
}

func TestFootprintFilesAreNotEditedByTheAgent(t *testing.T) {
	p := newProject(t)
	code, _, errOut := p.delta(hookEvent(t, "PreToolUse", "Write", filepath.Join(p.root, "deltas", "specs", "billing.files"), ""), "hook")
	if code != 2 || !strings.Contains(errOut, "knowledge base") {
		t.Fatalf("a write to the footprint = %d %q, want a block", code, errOut)
	}
}

func TestABuildRecordsManifestsButNotLockfilesOrBinaries(t *testing.T) {
	p := gitProject(t)
	p.write("go.mod", "module jobs\n")
	p.write("go.sum", "h1:one\n")
	p.write("deltas/specs/billing.spec.md", spec("billing", claimOf("INV-01", "none")))
	commitAll(t, p, "chore(repo): add the billing slice")

	if code, _, errOut := p.delta("", "ready", "billing"); code != 0 {
		t.Fatalf("ready = %d %q", code, errOut)
	}
	p.write("src/billing/total.go", "package billing\n")
	p.write("go.mod", "module jobs\n\nrequire example.com/money v1.0.0\n")
	p.write("go.sum", "h1:two\n")
	p.write("docs/diagram.png", "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	p.write("docs/diagram.svg", "<svg></svg>")
	code, out, errOut := p.delta("", "built", "billing")
	if code != 0 || errOut != "" || !strings.Contains(out, "billing footprint: 3 files") || !strings.Contains(out, "A person can remove what is not billing's from deltas/specs/billing.files.") {
		t.Fatalf("built = %d %q %q", code, out, errOut)
	}
	if got := p.read("deltas/specs/billing.files"); got != "docs/diagram.svg\ngo.mod\nsrc/billing/total.go\n" {
		t.Fatalf("footprint = %q; want the manifest and the text asset, and no lockfile or binary", got)
	}
}

func TestStopIgnoresLockfilesAndBinariesButNotManifests(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", built("billing", claimOf("INV-01", "test: git --version")))
	p.write("billing_test.go", "// INV-01\n")
	p.write("go.mod", "module jobs\n")
	p.write("go.sum", "h1:one\n")
	commitAll(t, p, "chore(repo): add the billing slice")
	if code, out, _ := p.delta("", "check"); code != 0 || !strings.Contains(out, "billing is verified") {
		t.Fatalf("check = %d %q", code, out)
	}
	commitAll(t, p, "chore(repo): verify billing")
	startSession(t, p)

	p.write("go.sum", "h1:two\n")
	p.write("assets/logo.png", "\x89PNG\x00\x00")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 0 || errOut != "" {
		t.Fatalf("a lockfile and an image: Stop = %d %q, want it allowed", code, errOut)
	}
	if code, out, _ := p.delta("", "stale"); code != 0 || out != "" {
		t.Fatalf("stale = %d %q; a lockfile and an image are not drift", code, out)
	}
	p.write("go.mod", "module jobs\n\nrequire example.com/money v1.0.0\n")
	if code, _, errOut := p.delta(stopEvent(t, false), "hook"); code != 2 || !strings.Contains(errOut, "delta check billing") {
		t.Fatalf("a manifest: Stop = %d %q, want a block that asks for the check", code, errOut)
	}
}

func TestTestingConfSaysWhatNeverCountsAsCode(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/testing.conf", "ignore=gen/**\n")
	p.write("deltas/specs/billing.spec.md", spec("billing", claimOf("INV-01", "none")))
	p.write("go.sum", "h1:one\n")
	commitAll(t, p, "chore(repo): add the billing slice")

	if code, _, errOut := p.delta("", "ready", "billing"); code != 0 {
		t.Fatalf("ready = %d %q", code, errOut)
	}
	p.write("gen/code.go", "package gen\n")
	p.write("go.sum", "h1:two\n")
	p.write("src/billing/total.go", "package billing\n")
	if code, _, errOut := p.delta("", "built", "billing"); code != 0 {
		t.Fatalf("built = %d %q", code, errOut)
	}
	if got := p.read("deltas/specs/billing.files"); got != "go.sum\nsrc/billing/total.go\n" {
		t.Fatalf("footprint = %q; the project's own patterns replace the lockfile defaults", got)
	}
}

func amended(slice string) string {
	return strings.Replace(spec(slice, "- INV-01 [test: git --version]\n  When x, z.\n"), "status: approved", "status: draft", 1)
}

func TestAnAmendedContractNeedsNoRebuildWhenItsTestsStillHold(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", spec("billing", claimOf("INV-01", "test: git --version")))
	p.write("billing_test.go", "// INV-01\n")
	p.write("src/billing/total.go", "package billing\n")
	commitAll(t, p, "chore(repo): add the billing slice")
	if code, _, errOut := p.delta("", "ready", "billing"); code != 0 {
		t.Fatalf("ready = %d %q", code, errOut)
	}
	p.write("src/billing/total.go", "package billing\n\nfunc Total() {}\n")
	if code, _, errOut := p.delta("", "built", "billing"); code != 0 {
		t.Fatalf("built = %d %q", code, errOut)
	}
	if code, out, _ := p.delta("", "check"); code != 0 || !strings.Contains(out, "billing is verified") {
		t.Fatalf("check = %d %q", code, out)
	}

	code, out, errOut := p.delta(amended("billing"), "spec", "write", "billing")
	if code != 0 || !strings.Contains(out, "wrote") || !strings.Contains(errOut, "draft again") {
		t.Fatalf("spec write = %d %q %q; want the amendment saved as a draft", code, out, errOut)
	}
	if code, _, errOut := p.delta("", "approve", "billing"); code != 0 {
		t.Fatalf("approve = %d %q", code, errOut)
	}
	if got := p.read("deltas/specs/billing.spec.md"); !strings.Contains(got, "status: approved") || !strings.Contains(got, "When x, z.") {
		t.Fatalf("spec = %q", got)
	}

	if code, out, _ := p.delta("", "check", "billing"); code != 0 || !strings.Contains(out, "billing is verified") {
		t.Fatalf("check = %d %q; want the amended slice verified without a rebuild", code, out)
	}
	if got := p.read("deltas/specs/billing.spec.md"); !strings.Contains(got, "status: verified") {
		t.Fatalf("spec = %q", got)
	}
	if got := p.read("deltas/specs/billing.files"); got != "src/billing/total.go\n" {
		t.Fatalf("footprint = %q, want it kept through the amendment", got)
	}
}

func TestAnApprovedSliceThatWasNeverBuiltStillNeedsBuilder(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/specs/billing.spec.md", spec("billing", claimOf("INV-01", "test: git --version")))
	p.write("billing_test.go", "// INV-01\n")
	commitAll(t, p, "chore(repo): add the billing slice")
	if code, out, _ := p.delta("", "check", "billing"); code != 0 || strings.Contains(out, "is verified") {
		t.Fatalf("check = %d %q; nothing was built, so nothing is verified", code, out)
	}
	if got := p.read("deltas/specs/billing.spec.md"); !strings.Contains(got, "status: approved") {
		t.Fatalf("spec = %q", got)
	}
}
