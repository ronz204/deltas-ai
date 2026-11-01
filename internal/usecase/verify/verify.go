// Package verify checks that specs are well formed, agree with the tests, and hold.
package verify

import (
	"context"

	"github.com/ronz204/deltas-ai/internal/domain/claim"
	"github.com/ronz204/deltas-ai/internal/domain/diag"
	"github.com/ronz204/deltas-ai/internal/domain/lint"
	"github.com/ronz204/deltas-ai/internal/domain/proof"
	"github.com/ronz204/deltas-ai/internal/domain/spec"
)

// Specs is where specs are read from.
type Specs interface {
	List(ctx context.Context) ([]string, error)
	Load(ctx context.Context, slice string) ([]byte, error)
	Path(slice string) string
}

// Tests gives the project's test files.
type Tests interface {
	TestFiles(ctx context.Context) ([]lint.File, error)
}

// Runner runs a claim's command and returns its exit code. An error means it did not finish.
type Runner interface {
	Run(ctx context.Context, command string) (exit int, err error)
}

// Checker verifies the specs of a project.
type Checker struct {
	specs Specs
	tests Tests
	run   Runner
}

func New(specs Specs, tests Tests, run Runner) *Checker {
	return &Checker{specs: specs, tests: tests, run: run}
}

type loaded struct {
	slice string
	file  string
	spec  spec.Spec
	diags []diag.Diagnostic
}

// load parses the named slice, or every slice when only is empty.
func (v *Checker) load(ctx context.Context, only string) ([]loaded, error) {
	names := []string{only}
	if only == "" {
		var err error
		if names, err = v.specs.List(ctx); err != nil {
			return nil, err
		}
	}
	out := make([]loaded, 0, len(names))
	for _, name := range names {
		src, err := v.specs.Load(ctx, name)
		if err != nil {
			return nil, err
		}
		file := v.specs.Path(name)
		sp, ds := spec.ParseSlice(file, name, src)
		out = append(out, loaded{slice: name, file: file, spec: sp, diags: ds})
	}
	return out, nil
}

func problems(ls []loaded) []diag.Diagnostic {
	var out []diag.Diagnostic
	for _, l := range ls {
		out = append(out, l.diags...)
	}
	return out
}

// SpecLint reports what is wrong with the structure of a spec, or of every spec when only is empty.
func (v *Checker) SpecLint(ctx context.Context, only string) ([]diag.Diagnostic, error) {
	ls, err := v.load(ctx, only)
	return problems(ls), err
}

// Lint is SpecLint over every spec plus the cross-check of specs against test files. A project
// whose specs have problems gets only those: a broken contract cannot be crossed.
func (v *Checker) Lint(ctx context.Context) ([]diag.Diagnostic, error) {
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
	slices := make([]lint.Slice, len(ls))
	for i, l := range ls {
		slices[i] = lint.Slice{File: l.file, Spec: l.spec}
	}
	return lint.Check(slices, files), nil
}

// Result is the verdict on one claim.
type Result struct {
	Slice   string
	ID      claim.ID
	Outcome proof.Outcome
}

// Report is what Check found: problems with the specs, and a verdict per claim of the specs
// that are sound.
type Report struct {
	Diags   []diag.Diagnostic
	Results []Result
}

// Check runs the command of every test claim of the named slice, or of every slice when only
// is empty, and judges each claim. A slice whose spec has problems is not run, only reported.
func (v *Checker) Check(ctx context.Context, only string) (Report, error) {
	ls, err := v.load(ctx, only)
	if err != nil {
		return Report{}, err
	}
	var files []lint.File
	if hasTestClaims(ls) {
		if files, err = v.tests.TestFiles(ctx); err != nil {
			return Report{}, err
		}
	}
	var rep Report
	for _, l := range ls {
		if len(l.diags) > 0 {
			rep.Diags = append(rep.Diags, l.diags...)
			continue
		}
		for _, c := range l.spec.Claims {
			if err := ctx.Err(); err != nil {
				return rep, err
			}
			rep.Results = append(rep.Results, Result{Slice: l.slice, ID: c.ID, Outcome: v.judge(ctx, c, files)})
		}
	}
	return rep, nil
}

func hasTestClaims(ls []loaded) bool {
	for _, l := range ls {
		if len(l.diags) > 0 {
			continue
		}
		for _, c := range l.spec.Claims {
			if c.Tag.Kind == claim.Test {
				return true
			}
		}
	}
	return false
}

// judge runs a test claim's command and decides the claim. A command that did not finish says
// nothing about the claim, so it leaves it Unverified, with the reason.
func (v *Checker) judge(ctx context.Context, c claim.Claim, files []lint.File) proof.Outcome {
	if c.Tag.Kind != claim.Test {
		return proof.Judge(c.Tag, 0, false)
	}
	exit, err := v.run.Run(ctx, c.Tag.Arg)
	if err != nil {
		return proof.Outcome{Verdict: proof.Unverified, Reason: "could not run the command: " + err.Error()}
	}
	return proof.Judge(c.Tag, exit, lint.Cited(c.ID, c.Tag.Arg, files))
}
