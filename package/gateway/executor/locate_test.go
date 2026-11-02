package executor

import (
	"os"
	"testing"
	"time"
)

func TestLocateFindsTheRunningProgramAndTellsItIsTheSameOne(t *testing.T) {
	r := New(t.TempDir(), time.Minute, nil)
	path, self := r.Locate(os.Args[0])
	if path == "" || !self {
		t.Fatalf("Locate(%q) = %q, %v; want this very test binary", os.Args[0], path, self)
	}
}

func TestLocateFindsAnotherProgramButNotAsItself(t *testing.T) {
	r := New(t.TempDir(), time.Minute, nil)
	path, self := r.Locate("go")
	if path == "" || self {
		t.Fatalf("Locate(go) = %q, %v; want the go tool, which is not this program", path, self)
	}
}

func TestLocateOfAMissingProgramIsEmpty(t *testing.T) {
	if path, self := New(t.TempDir(), time.Minute, nil).Locate("definitely-not-a-command-xyz"); path != "" || self {
		t.Fatalf("Locate = %q, %v; want nothing", path, self)
	}
}
