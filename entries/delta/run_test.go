package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelperProcess is the command the claims run: this test binary again, told by its argument to pass or fail.
func TestHelperProcess(t *testing.T) {
	if flag.Arg(0) == "fail" {
		os.Exit(1)
	}
}

// cmd is a claim command that exits 0 (pass) or 1 (fail); extra words are ignored by the helper but can name a path.
func cmd(mode string, extra ...string) string {
	return strings.Join(append([]string{fmt.Sprintf(`"%s" -test.run=^TestHelperProcess$ %s`, os.Args[0], mode)}, extra...), " ")
}

type project struct {
	t    *testing.T
	root string
}

func newProject(t *testing.T) *project {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return &project{t: t, root: root}
}

func (p *project) write(rel, text string) {
	p.t.Helper()
	path := filepath.Join(p.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		p.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		p.t.Fatal(err)
	}
}

func (p *project) read(rel string) string {
	p.t.Helper()
	b, err := os.ReadFile(filepath.Join(p.root, filepath.FromSlash(rel)))
	if err != nil {
		p.t.Fatal(err)
	}
	return string(b)
}

func (p *project) delta(stdin string, args ...string) (code int, out, errOut string) {
	return p.deltaIn(p.root, stdin, args...)
}

func (p *project) deltaIn(cwd, stdin string, args ...string) (code int, out, errOut string) {
	var o, e bytes.Buffer
	code = run(context.Background(), args, strings.NewReader(stdin), &o, &e, cwd)
	return code, o.String(), e.String()
}

func spec(slice string, claims ...string) string {
	return "---\nformat: 1\nslice: " + slice + "\nstatus: approved\n---\n" + strings.Join(claims, "")
}

func claimOf(id, tag string) string { return "- " + id + " [" + tag + "]\n  When x, y.\n" }

func verdict(out, slice, id string) string {
	for _, line := range strings.Split(out, "\n") {
		if f := strings.Fields(line); len(f) >= 3 && f[0] == slice && f[1] == id {
			return f[2]
		}
	}
	return ""
}

func TestSpecNewLintAndWrite(t *testing.T) {
	p := newProject(t)
	if code, out, _ := p.delta("", "spec", "new", "billing"); code != 0 || !strings.Contains(out, "created deltas/billing.spec.md") {
		t.Fatalf("spec new = %d %q", code, out)
	}
	if !strings.Contains(p.read("deltas/billing.spec.md"), "status: draft") {
		t.Fatal("a new spec must start as a draft")
	}
	if code, _, errOut := p.delta("", "spec", "new", "billing"); code != 2 || !strings.Contains(errOut, "already exists") {
		t.Fatalf("a second spec new = %d %q", code, errOut)
	}
	if code, _, _ := p.delta("", "spec", "new", "Bad_Name"); code != 2 {
		t.Fatalf("a bad slice name = %d, want 2", code)
	}
	if code, _, errOut := p.delta("", "spec", "lint"); code != 0 || errOut != "" {
		t.Fatalf("the template should lint clean: %d %q", code, errOut)
	}

	good := spec("billing", claimOf("INV-01", "none"))
	if code, out, _ := p.delta(good, "spec", "write", "billing"); code != 0 || !strings.Contains(out, "wrote") {
		t.Fatalf("spec write = %d %q", code, out)
	}
	if p.read("deltas/billing.spec.md") != good {
		t.Fatal("the written spec is not what was sent")
	}

	for name, bad := range map[string]string{
		"no frontmatter": "just text",
		"wrong slice":    spec("payments", claimOf("INV-01", "none")),
		"bad tag":        spec("billing", claimOf("INV-01", "bogus")),
	} {
		code, _, errOut := p.delta(bad, "spec", "write", "billing")
		if code != 1 || !strings.Contains(errOut, "nothing was written") {
			t.Fatalf("%s: spec write = %d %q, want 1", name, code, errOut)
		}
		if p.read("deltas/billing.spec.md") != good {
			t.Fatalf("%s: an invalid write changed the spec", name)
		}
	}
}

func TestCheckGivesEveryKindOfVerdict(t *testing.T) {
	p := newProject(t)
	p.write("deltas/billing.spec.md", spec("billing",
		claimOf("INV-01", "test: "+cmd("pass")),
		claimOf("INV-02", "test: "+cmd("fail")),
		claimOf("INV-03", "test: "+cmd("pass")),
		claimOf("INV-04", "manual: read the invoice"),
		claimOf("INV-05", "none"),
		claimOf("INV-06", "test: definitely-not-a-command-xyz"),
	))
	p.write("billing/total_test.go", "// INV-01\n// INV-02\n// INV-06\n")

	code, out, _ := p.delta("", "check")
	if code != 1 {
		t.Fatalf("a violated claim must exit 1, got %d\n%s", code, out)
	}
	want := map[string]string{"INV-01": "Holds", "INV-02": "Violated", "INV-03": "Unverified", "INV-04": "Unverified", "INV-05": "Unverified", "INV-06": "Unverified"}
	for id, v := range want {
		if got := verdict(out, "billing", id); got != v {
			t.Errorf("%s = %q, want %q\n%s", id, got, v, out)
		}
	}
	if !strings.Contains(out, "1 holds, 1 violated, 4 unverified") {
		t.Errorf("summary missing:\n%s", out)
	}
	if !strings.Contains(out, "could not run the command") {
		t.Errorf("an unrunnable command must say why:\n%s", out)
	}
}

func TestCheckExitsZeroWhenNothingIsViolated(t *testing.T) {
	p := newProject(t)
	p.write("deltas/billing.spec.md", spec("billing", claimOf("INV-01", "test: "+cmd("pass")), claimOf("INV-02", "none")))
	p.write("billing/total_test.go", "// INV-01\n")
	if code, out, _ := p.delta("", "check"); code != 0 {
		t.Fatalf("exit %d with nothing violated\n%s", code, out)
	}
}

func TestACitationOnlyCountsUnderTheCommandsPaths(t *testing.T) {
	p := newProject(t)
	p.write("deltas/billing.spec.md", spec("billing", claimOf("INV-01", "test: "+cmd("pass", "./billing"))))
	p.write("deltas/payments.spec.md", spec("payments", claimOf("INV-01", "test: "+cmd("pass", "./payments"))))
	p.write("billing/total_test.go", "// INV-01\n")
	p.write("payments/pay_test.go", "// nothing cited\n")

	_, out, _ := p.delta("", "check")
	if got := verdict(out, "billing", "INV-01"); got != "Holds" {
		t.Errorf("billing INV-01 = %q, want Holds\n%s", got, out)
	}
	if got := verdict(out, "payments", "INV-01"); got != "Unverified" {
		t.Errorf("payments INV-01 = %q, want Unverified: billing's test must not vouch for it\n%s", got, out)
	}
}

func TestLinkThenCheckClosesTheLoop(t *testing.T) {
	p := newProject(t)
	p.write("deltas/billing.spec.md", spec("billing", claimOf("INV-01", "none")))
	p.write("billing/total_test.go", "func TestINV01(t *testing.T) {}\n")
	if code, _, _ := p.delta("", "check"); code != 0 {
		t.Fatal("a spec with only a none claim must check clean")
	}
	if code, out, errOut := p.delta("", "link", "billing", "INV-01", cmd("pass")); code != 0 {
		t.Fatalf("link = %d %q %q", code, out, errOut)
	}
	if !strings.Contains(p.read("deltas/billing.spec.md"), "- INV-01 [test: ") {
		t.Fatal("link did not change the tag")
	}
	code, out, _ := p.delta("", "check")
	if code != 0 || verdict(out, "billing", "INV-01") != "Holds" {
		t.Fatalf("after link: exit %d\n%s", code, out)
	}
	if code, _, _ := p.delta("", "link", "billing", "INV-09", cmd("pass")); code != 2 {
		t.Fatalf("linking an unknown claim = %d, want 2", code)
	}
}

func TestLint(t *testing.T) {
	p := newProject(t)
	p.write("deltas/billing.spec.md", spec("billing", claimOf("INV-01", "test: "+cmd("pass"))))
	code, _, errOut := p.delta("", "lint")
	if code != 1 || !strings.Contains(errOut, "LINT-NO-CITE") {
		t.Fatalf("an uncited test claim = %d %q, want 1 and LINT-NO-CITE", code, errOut)
	}
	p.write("billing/total_test.go", "// INV-01\nt.Skip(\"later\")\n// INV-77 is gone\n")
	code, _, errOut = p.delta("", "lint")
	if code != 0 {
		t.Fatalf("warnings alone must not fail: %d %q", code, errOut)
	}
	for _, want := range []string{"warning LINT-SKIP", "warning LINT-UNKNOWN-ID"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("missing %q in %q", want, errOut)
		}
	}
}

func TestStatus(t *testing.T) {
	p := newProject(t)
	if code, out, _ := p.delta("", "status"); code != 0 || !strings.Contains(out, "delta spec new") {
		t.Fatalf("an empty project = %d %q", code, out)
	}
	p.write("deltas/billing.spec.md", spec("billing", claimOf("INV-01", "test: x"), claimOf("INV-02", "none")))
	p.write("deltas/broken.spec.md", "not a spec")
	_, out, _ := p.delta("", "status")
	for _, want := range []string{"SLICE", "billing", "approved", "broken"} {
		if !strings.Contains(out, want) {
			t.Errorf("status lacks %q:\n%s", want, out)
		}
	}
	for _, line := range strings.Split(out, "\n") {
		if f := strings.Fields(line); len(f) == 7 && f[0] == "billing" && strings.Join(f[2:], " ") != "2 1 0 1 0" {
			t.Errorf("billing counts are %v, want 2 claims, 1 test, 0 manual, 1 none, 0 issues", f[2:])
		}
	}
}

func TestCommitMsg(t *testing.T) {
	p := newProject(t)
	p.write("good", "feat(domain): add spec parser and claim rules\n# comment\n")
	p.write("bad", "Updated stuff.\n")
	if code, _, errOut := p.delta("", "commit-msg", filepath.Join(p.root, "good")); code != 0 {
		t.Fatalf("good message = %d %q", code, errOut)
	}
	code, _, errOut := p.delta("", "commit-msg", filepath.Join(p.root, "bad"))
	if code != 1 || !strings.Contains(errOut, "COMMIT-") {
		t.Fatalf("bad message = %d %q", code, errOut)
	}
	if code, _, _ := p.delta("", "commit-msg", filepath.Join(p.root, "missing")); code != 2 {
		t.Fatalf("a missing message file = %d, want 2", code)
	}
}

func TestRunsFromASubdirectory(t *testing.T) {
	p := newProject(t)
	p.write("deltas/billing.spec.md", spec("billing", claimOf("INV-01", "none")))
	p.write("sub/dir/keep", "")
	code, out, _ := p.deltaIn(filepath.Join(p.root, "sub", "dir"), "", "status")
	if code != 0 || !strings.Contains(out, "billing") {
		t.Fatalf("status from a subdirectory = %d %q", code, out)
	}
}

func TestUsageErrorsExitTwo(t *testing.T) {
	p := newProject(t)
	for _, args := range [][]string{{"nope"}, {"link", "only-one"}, {"spec", "new"}, {"check", "a", "b"}, {"lint", "x"}} {
		code, _, errOut := p.delta("", args...)
		if code != 2 || !strings.HasPrefix(errOut, "delta: ") {
			t.Errorf("%v = %d %q, want 2 and a delta: message", args, code, errOut)
		}
	}
}
