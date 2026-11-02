package main

import (
	"strings"
	"testing"
)

// A project whose tests are Ruby specs: testing.conf says so, and a citation in one of them makes a claim hold.
func TestTestingConfChoosesWhichFilesAreTests(t *testing.T) {
	p := newProject(t)
	p.write("deltas/specs/billing.spec.md", spec("billing", claimOf("INV-01", "test: "+command("pass"))))
	p.write("spec/billing_spec.rb", "# INV-01\n")

	if code, _, errOut := p.delta("", "lint"); code != 1 || !strings.Contains(errOut, "LINT-NO-CITE") {
		t.Fatalf("lint = %d %q; without testing.conf a _spec.rb file is not a test", code, errOut)
	}
	p.write("deltas/testing.conf", "match=*_spec.rb\n")
	if code, _, errOut := p.delta("", "lint"); code != 0 {
		t.Fatalf("lint = %d %q; with testing.conf the Ruby spec cites INV-01", code, errOut)
	}
	if code, out, _ := p.delta("", "check"); code != 0 || verdict(out, "billing", "INV-01") != "Holds" {
		t.Fatalf("check = %d %q", code, out)
	}
}

// A folder named in skip is never searched, so a test file inside it cites nothing.
func TestTestingConfChoosesWhereNotToLookForTests(t *testing.T) {
	p := newProject(t)
	p.write("deltas/specs/billing.spec.md", spec("billing", claimOf("INV-01", "test: "+command("pass"))))
	p.write("dist/billing_test.go", "// INV-01\n")

	if code, _, errOut := p.delta("", "lint"); code != 0 {
		t.Fatalf("lint = %d %q; by default dist/ is searched", code, errOut)
	}
	p.write("deltas/testing.conf", "skip=dist\n")
	if code, _, errOut := p.delta("", "lint"); code != 1 || !strings.Contains(errOut, "LINT-NO-CITE") {
		t.Fatalf("lint = %d %q; with skip=dist the test in dist/ is not seen", code, errOut)
	}
}

func TestBudgetConfSetsTheBudgetCeilings(t *testing.T) {
	p := newProject(t)
	p.write(".claude/rules/always.md", "# Rule\n"+strings.Repeat("a line\n", 10))
	if code, _, _ := p.delta("", "budget"); code != 0 {
		t.Fatalf("budget = %d with the default ceilings", code)
	}
	p.write("deltas/budget.conf", "rule_always_lines=5\n")
	code, _, errOut := p.delta("", "budget")
	if code != 1 || !strings.Contains(errOut, "the ceiling for a rule is 5") {
		t.Fatalf("budget = %d %q, want budget.conf's ceiling", code, errOut)
	}
}

func TestStatusWarnsAboutProblemsInTheConfigFiles(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/budget.conf", "colour=blue\n")
	p.write("deltas/testing.conf", "size=3\n")
	_, _, errOut := p.delta("", "status")
	for _, want := range []string{`deltas/budget.conf:1: unknown key "colour"`, `deltas/testing.conf:1: unknown key "size"`} {
		if !strings.Contains(errOut, want) {
			t.Errorf("status warnings lack %q:\n%s", want, errOut)
		}
	}
}
