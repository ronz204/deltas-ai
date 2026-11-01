package measure

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/budget"
)

var ctx = context.Background()

type fakeFiles struct {
	conf    *string
	kb      []budget.File
	confErr error
	globbed []string
}

func (f *fakeFiles) ReadFile(_ context.Context, rel string) ([]byte, error) {
	if rel != ".claude/budget.conf" {
		return nil, fmt.Errorf("unexpected read of %s", rel)
	}
	if f.confErr != nil {
		return nil, f.confErr
	}
	if f.conf == nil {
		return nil, fmt.Errorf("read %s: %w", rel, fs.ErrNotExist)
	}
	return []byte(*f.conf), nil
}

func (f *fakeFiles) Glob(_ context.Context, patterns []string) ([]budget.File, error) {
	f.globbed = patterns
	return f.kb, nil
}

func rule(lines int) budget.File {
	return budget.File{Path: ".claude/rules/a.md", Text: strings.Repeat("x\n", lines)}
}

func TestAlwaysUsesTheDefaultsWithoutAConfFile(t *testing.T) {
	f := &fakeFiles{kb: []budget.File{rule(51)}}
	rep, err := New(f).Always(ctx)
	if err != nil || rep.Limit != 3500 || len(rep.Findings()) != 1 {
		t.Fatalf("report = %+v, %v; want the default ceilings and one rule over 50 lines", rep, err)
	}
	if len(f.globbed) == 0 || f.globbed[0] != "CLAUDE.md" {
		t.Fatalf("it looked for %v, want the patterns of the always-loaded files", f.globbed)
	}
}

func TestAlwaysReadsTheCeilingsFromTheConfFile(t *testing.T) {
	conf := "rule_always_lines=100\nalways_tokens=10\n"
	rep, err := New(&fakeFiles{conf: &conf, kb: []budget.File{rule(51)}}).Always(ctx)
	if err != nil || rep.Limit != 10 {
		t.Fatalf("report = %+v, %v", rep, err)
	}
	ds := rep.Findings()
	if len(ds) != 1 || ds[0].Code != budget.CodeTotal {
		t.Fatalf("findings = %v, want only the total (the rule is under its raised ceiling)", ds)
	}
}

func TestAlwaysSurfacesAReadFailure(t *testing.T) {
	if _, err := New(&fakeFiles{confErr: errors.New("disk on fire")}).Always(ctx); err == nil {
		t.Fatal("a conf file that cannot be read was taken as missing")
	}
}
