package terminal

import (
	"errors"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/package/domain/linking"
)

const draftHead = "---\nformat: 1\nslice: billing\nstatus: draft\n---\n"

func TestApproveThenBuiltThenCheckWalkTheCycle(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = draftHead + claim("INV-01", "test: go test ./a")
	for _, step := range []struct{ args, want string }{
		{"approve billing", "billing is approved\n"},
		{"built billing", "billing is built\n"},
	} {
		out, _, err := m.run("", strings.Fields(step.args)...)
		if err != nil || out != step.want {
			t.Fatalf("delta %s = %q, %v", step.args, out, err)
		}
	}
	if !strings.Contains(m.specs["billing"], "status: built") {
		t.Fatalf("stored %q", m.specs["billing"])
	}
}

func TestCheckSaysWhichSlicesItVerified(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = strings.Replace(head, "approved", "built", 1) + claim("INV-01", "test: go test ./a")
	m.files = []linking.File{{Path: "a/a_test.go", Text: "// INV-01"}}
	out, _, err := m.run("", "check")
	if err != nil || !strings.HasSuffix(out, "billing is verified\n") || !strings.Contains(m.specs["billing"], "status: verified") {
		t.Fatalf("out = %q, err = %v, stored %q", out, err, m.specs["billing"])
	}
}

func TestApproveRefusesWhatItCannotApprove(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = head + claim("INV-01", "none")
	if _, _, err := m.run("", "approve", "billing"); err == nil || errors.Is(err, ErrFound) {
		t.Fatalf("approving an approved slice = %v, want a plain error", err)
	}
	m.specs["broken"] = "no frontmatter"
	_, errOut, err := m.run("", "approve", "broken")
	if !errors.Is(err, ErrFound) || !strings.Contains(errOut, "SPEC-FRONTMATTER") || !strings.Contains(errOut, "nothing was changed") {
		t.Fatalf("err = %v, stderr = %q", err, errOut)
	}
}

func TestReadyIsASilentGateForForge(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = head + claim("INV-01", "none")
	out, errOut, err := m.run("", "ready", "billing")
	if err != nil || out != "billing is ready to build\n" || errOut != "" {
		t.Fatalf("an approved slice: %q %q %v", out, errOut, err)
	}
	m.specs["billing"] = draftHead + claim("INV-01", "none")
	out, errOut, err = m.run("", "ready", "billing")
	if !errors.Is(err, ErrFound) || out != "" || !strings.Contains(errOut, "delta approve billing") {
		t.Fatalf("a draft: %q %q %v", out, errOut, err)
	}
	if _, _, err := m.run("", "ready", "ghost"); err == nil || errors.Is(err, ErrFound) {
		t.Fatalf("a missing spec = %v, want a plain error", err)
	}
}

func TestSpecWriteWarnsWhenAChangedContractFallsBackToDraft(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = head + claim("INV-01", "none")
	changed := head + strings.Replace(claim("INV-01", "none"), "When x, y.", "When x, z.", 1)
	out, errOut, err := m.run(changed, "spec", "write", "billing")
	if err != nil || out != "wrote the spec of billing\n" || !strings.Contains(errOut, "warning AUTHOR-REAPPROVE") {
		t.Fatalf("out = %q, errOut = %q, err = %v", out, errOut, err)
	}
	if !strings.Contains(m.specs["billing"], "status: draft") {
		t.Fatalf("stored %q", m.specs["billing"])
	}
}
