package sentinel

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/judgment"
	"github.com/ronz204/deltas-ai/package/domain/linking"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

func (g *Guard) stop(ctx context.Context, e Event) (Decision, error) {
	if e.StopActive || g.evidence == nil || e.Session == "" {
		return Decision{}, nil
	}
	touched, known, err := g.evidence.Touched(ctx, e.Session)
	if err != nil {
		return Decision{}, err
	}
	touched = slices.DeleteFunc(touched, func(file string) bool { return !linking.Code(file) })
	if !known || len(touched) == 0 {
		return Decision{}, nil
	}
	rows, err := g.slices.Report(ctx)
	if err != nil {
		return Decision{}, err
	}
	var names []string
	scopes := map[string]linking.Scope{}
	for _, r := range rows {
		if r.Issues == 0 && r.Test > 0 && wasBuilt(r.Status) && slices.ContainsFunc(touched, r.Scope.Covers) {
			names = append(names, r.Slice)
			scopes[r.Slice] = r.Scope
		}
	}
	if len(names) == 0 {
		return Decision{}, nil
	}
	by, err := g.evidence.Standings(ctx, scopes)
	if err != nil {
		return Decision{}, err
	}
	var unchecked, failing []string
	for _, n := range names {
		switch by[n] {
		case judgment.Unchecked:
			unchecked = append(unchecked, n)
		case judgment.Failing:
			failing = append(failing, n)
		}
	}
	var lines []string
	if len(unchecked) > 0 {
		lines = append(lines, fmt.Sprintf("Your changes touched code of %s: run delta check %s before you finish.", strings.Join(unchecked, ", "), strings.Join(unchecked, " ")))
	}
	if len(failing) > 0 {
		lines = append(lines, "delta check found violated claims for: "+strings.Join(failing, ", ")+". Fix the code; if the promise itself changed, tell the person, who decides it with /drafter.")
	}
	if len(lines) == 0 {
		return Decision{}, nil
	}
	return Decision{Block: true, Reason: strings.Join(lines, "\n")}, nil
}

func wasBuilt(s states.Status) bool {
	return s == states.Built || s == states.Verified || s == states.Stale
}
