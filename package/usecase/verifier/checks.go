package verifier

import (
	"context"

	"github.com/ronz204/deltas-ai/package/domain/claimset"
	"github.com/ronz204/deltas-ai/package/domain/findings"
	"github.com/ronz204/deltas-ai/package/domain/judgment"
	"github.com/ronz204/deltas-ai/package/domain/linking"
	"github.com/ronz204/deltas-ai/package/domain/specfile"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

type Result struct {
	Slice   string
	ID      claimset.ID
	Outcome judgment.Outcome
}

type Report struct {
	Diags    []findings.Diagnostic
	Results  []Result
	Verified []string
}

const (
	CodeRecord = "CHECK-RECORD"
	CodeStatus = "CHECK-STATUS"
)

func (v *Checker) Check(ctx context.Context, only string) (Report, error) {
	ls, err := v.load(ctx, only)
	if err != nil {
		return Report{}, err
	}
	var files []linking.File
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

func (v *Checker) checkSlice(ctx context.Context, l loaded, files []linking.File, rep *Report) (Seen, error) {
	var footprint []string
	if v.footprints != nil {
		var err error
		if footprint, err = v.footprints.Footprint(ctx, l.slice); err != nil {
			return Seen{}, err
		}
	}
	s := Seen{Slice: l.slice, Scope: linking.NewScope(l.file, l.spec.Claims, footprint)}
	var held, open int
	for _, c := range l.spec.Claims {
		if err := ctx.Err(); err != nil {
			return s, err
		}
		out := v.judge(ctx, c, files)
		rep.Results = append(rep.Results, Result{Slice: l.slice, ID: c.ID, Outcome: out})
		s.Violated = s.Violated || out.Verdict == judgment.Violated
		switch {
		case c.Tag.Method != claimset.Test:
		case out.Verdict == judgment.Holds:
			held++
		default:
			open++
		}
	}
	switch {
	case held == 0 || open > 0:
		return s, nil
	case l.spec.Status == states.Verified:
		s.Verified = true
		return s, nil
	case l.spec.Status != states.Built && l.spec.Status != states.Stale:
		return s, nil
	}
	if err := v.verified(ctx, l); err != nil {
		rep.Diags = append(rep.Diags, findings.Diagnostic{File: l.file, Code: CodeStatus, Warn: true,
			Message: "every test claim holds, but the slice could not be marked verified: " + err.Error()})
		return s, nil //nolint:nilerr // reported as a warning in the report, not as a failure of the check
	}
	rep.Verified = append(rep.Verified, l.slice)
	s.Verified = true
	return s, nil
}

func (v *Checker) verified(ctx context.Context, l loaded) error {
	src, err := v.specs.Load(ctx, l.slice)
	if err != nil {
		return err
	}
	edited, err := specfile.SetStatus(src, states.Verified)
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
			if c.Tag.Method == claimset.Test {
				return true
			}
		}
	}
	return false
}

func (v *Checker) judge(ctx context.Context, c claimset.Claim, files []linking.File) judgment.Outcome {
	if c.Tag.Method != claimset.Test {
		return judgment.Judge(c.Tag, 0, false)
	}
	exit, err := v.run.Run(ctx, c.Tag.Arg)
	if err != nil {
		return judgment.Outcome{Verdict: judgment.Unverified, Reason: "could not run the command: " + err.Error()}
	}
	return judgment.Judge(c.Tag, exit, linking.Cited(c.ID, c.Tag.Arg, files))
}
