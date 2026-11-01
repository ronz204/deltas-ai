// Package verify checks that specs are well formed, agree with the tests, and hold.
package verify

import (
	"context"

	"github.com/ronz204/deltas-ai/source/domain/assertion"
	"github.com/ronz204/deltas-ai/source/domain/findings"
	"github.com/ronz204/deltas-ai/source/domain/judgment"
	"github.com/ronz204/deltas-ai/source/domain/lifecycle"
	"github.com/ronz204/deltas-ai/source/domain/linkcheck"
	"github.com/ronz204/deltas-ai/source/domain/specfile"
)

type Specs interface {
	List(ctx context.Context) ([]string, error)
	Load(ctx context.Context, slice string) ([]byte, error)
	Save(ctx context.Context, slice string, content []byte) error
	Path(slice string) string
	// Resolve gives the full name of the slice a reference points to, or an error if it is ambiguous.
	Resolve(ctx context.Context, ref string) (string, error)
}

type Tests interface {
	TestFiles(ctx context.Context) ([]linkcheck.File, error)
	IsTest(rel string) bool
}

// Runner runs a claim's command and returns its exit code. An error means it did not finish.
type Runner interface {
	Run(ctx context.Context, command, dir string) (exit int, err error)
}

// Seen is what a check found for one slice: whether a claim was violated, and the part of the project it depends on.
type Seen struct {
	Slice    string
	Scope    linkcheck.Scope
	Violated bool
}

// Marks remembers what a check saw, so the Stop hook can tell a checked tree from an unchecked one.
type Marks interface {
	Record(ctx context.Context, seen []Seen) error
}

// Codes of the warnings of a check that ran but could not remember something.
const (
	CodeRecord = "CHECK-RECORD"
	CodeStatus = "CHECK-STATUS"
)

type Checker struct {
	specs  Specs
	tests  Tests
	run    Runner
	marks  Marks
	scopes map[string][]string
}

// New builds the checker. marks may be nil, and then checks are not remembered. scopes holds, for the slices
// that have them, the paths delta.conf says they depend on; any other slice depends on every code file.
func New(specs Specs, tests Tests, run Runner, marks Marks, scopes map[string][]string) *Checker {
	return &Checker{specs: specs, tests: tests, run: run, marks: marks, scopes: scopes}
}

// IsTest says whether a file of the project is a test file.
func (v *Checker) IsTest(rel string) bool { return v.tests.IsTest(rel) }

type loaded struct {
	slice string
	file  string
	spec  specfile.Spec
	diags []findings.Diagnostic
}

// load parses the named slice, or every slice when only is empty.
func (v *Checker) load(ctx context.Context, only string) ([]loaded, error) {
	var names []string
	var err error
	if only == "" {
		names, err = v.specs.List(ctx)
	} else {
		var name string
		name, err = v.specs.Resolve(ctx, only)
		names = []string{name}
	}
	if err != nil {
		return nil, err
	}
	out := make([]loaded, 0, len(names))
	for _, name := range names {
		src, err := v.specs.Load(ctx, name)
		if err != nil {
			return nil, err
		}
		file := v.specs.Path(name)
		sp, ds := specfile.ParseSlice(file, name, src)
		out = append(out, loaded{slice: name, file: file, spec: sp, diags: ds})
	}
	return out, nil
}

func problems(ls []loaded) []findings.Diagnostic {
	var out []findings.Diagnostic
	for _, l := range ls {
		out = append(out, l.diags...)
	}
	return out
}

func (v *Checker) SpecLint(ctx context.Context, only string) ([]findings.Diagnostic, error) {
	ls, err := v.load(ctx, only)
	return problems(ls), err
}

// Lint adds the cross-check with test files; a project whose specs have problems gets only those.
func (v *Checker) Lint(ctx context.Context) ([]findings.Diagnostic, error) {
	ls, err := v.load(ctx, "")
	if err != nil {
		return nil, err
	}
	if ds := problems(ls); len(ds) > 0 {
		return ds, nil
	}
	files, err := v.tests.TestFiles(ctx)
	if err != nil {
		return nil, err
	}
	slices := make([]linkcheck.Slice, len(ls))
	for i, l := range ls {
		slices[i] = linkcheck.Slice{File: l.file, Dir: specfile.Dir(l.slice), Spec: l.spec}
	}
	return linkcheck.Check(slices, files), nil
}

type Result struct {
	Slice   string
	ID      assertion.ID
	Outcome judgment.Outcome
}

type Report struct {
	Diags    []findings.Diagnostic
	Results  []Result
	Verified []string // slices this check moved from built to verified
}

// Check judges every claim; a slice whose spec has problems is only reported, not run. A built slice whose test
// claims all hold (and that has at least one) becomes verified; manual and none claims stay visible as debt.
func (v *Checker) Check(ctx context.Context, only string) (Report, error) {
	ls, err := v.load(ctx, only)
	if err != nil {
		return Report{}, err
	}
	var files []linkcheck.File
	if hasTestClaims(ls) {
		if files, err = v.tests.TestFiles(ctx); err != nil {
			return Report{}, err
		}
	}
	var rep Report
	var seen []Seen
	for _, l := range ls {
		if len(l.diags) > 0 {
			rep.Diags = append(rep.Diags, l.diags...)
			continue
		}
		s, err := v.checkSlice(ctx, l, files, &rep)
		if err != nil {
			return rep, err
		}
		seen = append(seen, s)
	}
	if v.marks != nil && len(seen) > 0 {
		if err := v.marks.Record(ctx, seen); err != nil {
			rep.Diags = append(rep.Diags, findings.Diagnostic{File: "delta check", Code: CodeRecord, Warn: true,
				Message: "the check could not be remembered, so the Stop hook will ask for it again: " + err.Error()})
		}
	}
	return rep, nil
}

// checkSlice judges the claims of one clean slice into rep, and verifies it if it is built and its test claims all hold.
func (v *Checker) checkSlice(ctx context.Context, l loaded, files []linkcheck.File, rep *Report) (Seen, error) {
	s := Seen{Slice: l.slice, Scope: linkcheck.NewScope(l.file, v.scopes[l.slice])}
	dir := specfile.Dir(l.slice)
	files = linkcheck.Under(files, dir)
	var held, open int // test claims that hold, and test claims that do not
	for _, c := range l.spec.Claims {
		if err := ctx.Err(); err != nil {
			return s, err
		}
		out := v.judge(ctx, c, files, dir)
		rep.Results = append(rep.Results, Result{Slice: l.slice, ID: c.ID, Outcome: out})
		s.Violated = s.Violated || out.Verdict == judgment.Violated
		switch {
		case c.Tag.Method != assertion.Test:
		case out.Verdict == judgment.Holds:
			held++
		default:
			open++
		}
	}
	if l.spec.Status != lifecycle.Built || held == 0 || open > 0 {
		return s, nil
	}
	if err := v.verified(ctx, l); err != nil {
		rep.Diags = append(rep.Diags, findings.Diagnostic{File: l.file, Code: CodeStatus, Warn: true,
			Message: "every test claim holds, but the slice could not be marked verified: " + err.Error()})
		return s, nil
	}
	rep.Verified = append(rep.Verified, l.slice)
	return s, nil
}

// verified moves a built slice to verified. It runs before the check is remembered, so the mark covers the edit.
func (v *Checker) verified(ctx context.Context, l loaded) error {
	src, err := v.specs.Load(ctx, l.slice)
	if err != nil {
		return err
	}
	edited, err := specfile.SetStatus(src, lifecycle.Verified)
	if err != nil {
		return err
	}
	return v.specs.Save(ctx, l.slice, edited)
}

func hasTestClaims(ls []loaded) bool {
	for _, l := range ls {
		if len(l.diags) > 0 {
			continue
		}
		for _, c := range l.spec.Claims {
			if c.Tag.Method == assertion.Test {
				return true
			}
		}
	}
	return false
}

func (v *Checker) judge(ctx context.Context, c assertion.Claim, files []linkcheck.File, dir string) judgment.Outcome {
	if c.Tag.Method != assertion.Test {
		return judgment.Judge(c.Tag, 0, false)
	}
	exit, err := v.run.Run(ctx, c.Tag.Arg, dir)
	if err != nil {
		return judgment.Outcome{Verdict: judgment.Unverified, Reason: "could not run the command: " + err.Error()}
	}
	return judgment.Judge(c.Tag, exit, linkcheck.Cited(c.ID, c.Tag.Arg, files))
}
