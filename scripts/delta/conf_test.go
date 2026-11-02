package main

import (
	"strings"
	"testing"
)

// A project whose tests are Ruby specs: deltas.config says so, and a citation in one of them makes a claim hold.
func TestDeltaConfChoosesWhichFilesAreTests(t *testing.T) {
	p := newProject(t)
	p.write("deltas/specs/billing.spec.md", spec("billing", claimOf("INV-01", "test: "+command("pass"))))
	p.write("spec/billing_spec.rb", "# INV-01\n")

	if code, _, errOut := p.delta("", "lint"); code != 1 || !strings.Contains(errOut, "LINT-NO-CITE") {
		t.Fatalf("lint = %d %q; without deltas.config a _spec.rb file is not a test", code, errOut)
	}
	p.write("deltas/deltas.config", "test_files=*_spec.rb\n")
	if code, _, errOut := p.delta("", "lint"); code != 0 {
		t.Fatalf("lint = %d %q; with deltas.config the Ruby spec cites INV-01", code, errOut)
	}
	if code, out, _ := p.delta("", "check"); code != 0 || verdict(out, "billing", "INV-01") != "Holds" {
		t.Fatalf("check = %d %q", code, out)
	}
}

func TestDeltaConfSetsTheBudgetCeilings(t *testing.T) {
	p := newProject(t)
	p.write(".claude/rules/always.md", "# Rule\n"+strings.Repeat("a line\n", 10))
	if code, _, _ := p.delta("", "budget"); code != 0 {
		t.Fatalf("budget = %d with the default ceilings", code)
	}
	p.write("deltas/deltas.config", "rule_always_lines=5\n")
	code, _, errOut := p.delta("", "budget")
	if code != 1 || !strings.Contains(errOut, "the ceiling for a rule is 5") {
		t.Fatalf("budget = %d %q, want deltas.config's ceiling", code, errOut)
	}
}

func TestTheOlderBudgetConfStillWorksWithoutDeltaConf(t *testing.T) {
	p := newProject(t)
	p.write(".claude/rules/always.md", "# Rule\n"+strings.Repeat("a line\n", 10))
	p.write(".claude/budget.conf", "rule_always_lines=5\nreference_lines=150\n")
	code, _, errOut := p.delta("", "budget")
	if code != 1 || !strings.Contains(errOut, "the ceiling for a rule is 5") {
		t.Fatalf("budget = %d %q, want the older file's ceiling", code, errOut)
	}
	p.write("deltas/deltas.config", "rule_always_lines=20\n")
	if code, _, errOut := p.delta("", "budget"); code != 0 {
		t.Fatalf("budget = %d %q, deltas.config should win over the older file", code, errOut)
	}
}

func TestStatusWarnsAboutDeltaConfProblemsAndMissingTools(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/deltas.config", "colour=blue\nrequired_tools=delta-no-such-tool-anywhere\n")
	_, _, errOut := p.delta("", "status")
	for _, want := range []string{`deltas/deltas.config:1: unknown key "colour"`, "delta-no-such-tool-anywhere is required by deltas.config"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("status warnings lack %q:\n%s", want, errOut)
		}
	}
}

func TestACommandTimeoutFromDeltaConfIsUsed(t *testing.T) {
	p := newProject(t)
	p.write("deltas/deltas.config", "command_timeout=1ns\n")
	p.write("deltas/specs/billing.spec.md", spec("billing", claimOf("INV-01", "test: "+command("pass"))))
	p.write("billing_test.go", "// INV-01\n")
	_, out, _ := p.delta("", "check")
	if verdict(out, "billing", "INV-01") != "Unverified" {
		t.Fatalf("check = %q; a command that cannot finish in 1ns is Unverified, not Holds", out)
	}
}
