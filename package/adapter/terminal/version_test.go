package terminal

import "testing"

func TestVersionPrintsTheLineGitDeltaCannotPrint(t *testing.T) {
	m := newMemory()
	out, _, err := m.run("", "version")
	if err != nil || out != "delta (deltas-ai) test\n" {
		t.Fatalf("out = %q, err = %v", out, err)
	}
}
