package verifier

import (
	"context"

	"github.com/ronz204/deltas-ai/package/domain/findings"
	"github.com/ronz204/deltas-ai/package/domain/linking"
	"github.com/ronz204/deltas-ai/package/domain/specfile"
)

type Specs interface {
	List(ctx context.Context) ([]string, error)
	Load(ctx context.Context, slice string) ([]byte, error)
	Save(ctx context.Context, slice string, content []byte) error
	Path(slice string) string
}

type Tests interface {
	TestFiles(ctx context.Context) ([]linking.File, error)
	IsTest(rel string) bool
}

type Runner interface {
	Run(ctx context.Context, command string) (exit int, err error)
}

type Seen struct {
	Slice    string
	Scope    linking.Scope
	Violated bool
	Verified bool
}

type Footprints interface {
	Footprint(ctx context.Context, slice string) ([]string, error)
}

type Marks interface {
	Record(ctx context.Context, seen []Seen) error
}

type Checker struct {
	specs      Specs
	tests      Tests
	run        Runner
	marks      Marks
	footprints Footprints
}

func New(specs Specs, tests Tests, run Runner, marks Marks) *Checker {
	return &Checker{specs: specs, tests: tests, run: run, marks: marks}
}

func (v *Checker) WithFootprints(f Footprints) *Checker {
	v.footprints = f
	return v
}

func (v *Checker) IsTest(rel string) bool { return v.tests.IsTest(rel) }

type loaded struct {
	slice string
	file  string
	spec  specfile.Spec
	diags []findings.Diagnostic
}

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
	slices := make([]linking.Slice, len(ls))
	for i, l := range ls {
		slices[i] = linking.Slice{File: l.file, Spec: l.spec}
	}
	return linking.Check(slices, files), nil
}
