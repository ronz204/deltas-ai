package measure

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/budget"
)

var ctx = context.Background()

type fakeFiles struct {
	kb      []budget.File
	err     error
	globbed []string
}

func (f *fakeFiles) Glob(_ context.Context, patterns []string) ([]budget.File, error) {
	f.globbed = patterns
	return f.kb, f.err
}

func rule(lines int) budget.File {
	return budget.File{Path: ".claude/rules/a.md", Text: strings.Repeat("x\n", lines)}
}

func TestAlwaysWeighsAgainstTheCeilingsItWasGiven(t *testing.T) {
	f := &fakeFiles{kb: []budget.File{rule(51)}}
	rep, err := New(f, budget.Defaults()).Always(ctx)
	if err != nil || rep.Limit != 3500 || len(rep.Findings()) != 1 {
		t.Fatalf("report = %+v, %v; want the default ceilings and one rule over 50 lines", rep, err)
	}
	if len(f.globbed) == 0 || f.globbed[0] != "CLAUDE.md" {
		t.Fatalf("it looked for %v, want the patterns of the always-loaded files", f.globbed)
	}

	raised := budget.Ceilings{Total: 10, Rule: 100}
	rep, err = New(&fakeFiles{kb: []budget.File{rule(51)}}, raised).Always(ctx)
	if err != nil || rep.Limit != 10 {
		t.Fatalf("report = %+v, %v", rep, err)
	}
	if ds := rep.Findings(); len(ds) != 1 || ds[0].Code != budget.CodeTotal {
		t.Fatalf("findings = %v, want only the total (the rule is under its raised ceiling)", ds)
	}
}

func TestAlwaysSurfacesAReadFailure(t *testing.T) {
	if _, err := New(&fakeFiles{err: errors.New("disk on fire")}, budget.Defaults()).Always(ctx); err == nil {
		t.Fatal("a failed read was taken as an empty project")
	}
}
