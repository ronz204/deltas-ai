package main

import (
	"encoding/json"
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
	p.write("deltas/billing.spec.md", spec("billing", claimOf("INV-01", "test: git --version")))
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
	p.write("deltas/billing.spec.md", spec("billing", claimOf("INV-01", "test: git no-such-subcommand")))
	p.write("billing_test.go", "// INV-01\n")
	if code, out, _ := p.delta("", "check"); code != 1 || verdict(out, "billing", "INV-01") != "Violated" {
		t.Fatalf("check = %d %q, want the claim violated", code, out)
	}
	code, _, errOut := p.delta(stopEvent(t, false), "hook")
	if code != 2 || !strings.Contains(errOut, "violated") {
		t.Fatalf("Stop = %d %q, want a block about violated claims", code, errOut)
	}
}
