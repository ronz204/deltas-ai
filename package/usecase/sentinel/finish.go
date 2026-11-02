package sentinel

import (
	"context"
	"fmt"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/judgment"
	"github.com/ronz204/deltas-ai/package/domain/linking"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

func (g *Guard) stop(ctx context.Context, e Event) (Decision, error) {
	if e.StopActive || g.evidence == nil {
		return Decision{}, nil
	}
	rows, err := g.slices.Report(ctx)
	if err != nil {
		return Decision{}, err
	}
	var names []string
	scopes := map[string]linking.Scope{}
	for _, r := range rows {
		if r.Issues == 0 && (r.Test > 0 || waitsForBuilder(r.Status)) {
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
	state := map[string]states.Status{}
	for _, r := range rows {
		state[r.Slice] = r.Status
	}
	var unchecked, failing, unbuilt []string
	for _, n := range names {
		switch {
		case by[n] == judgment.Passing:
		case waitsForBuilder(state[n]):
			unbuilt = append(unbuilt, n)
		case by[n] == judgment.Unchecked:
			unchecked = append(unchecked, n)
		case by[n] == judgment.Failing:
			failing = append(failing, n)
		}
	}
	var lines []string
	for _, n := range unbuilt {
		if state[n] == states.Stale {
			lines = append(lines, fmt.Sprintf("%s is stale and its code changed again: bring it up to date with /builder %s, which ends with delta built %s, or tell the user what is left.", n, n, n))
			continue
		}
		lines = append(lines, fmt.Sprintf("%s is approved but not built, and the code changed since: finish it with /builder %s, which ends with delta built %s, or tell the user what is left.", n, n, n))
	}
	if len(unchecked) > 0 {
		lines = append(lines, "delta check has not run on the current changes for: "+strings.Join(unchecked, ", ")+". Run it before you finish.")
	}
	if len(failing) > 0 {
		lines = append(lines, "delta check found violated claims for: "+strings.Join(failing, ", ")+". Fix them, or tell the user what is left.")
	}
	if len(lines) == 0 {
		return Decision{}, nil
	}
	return Decision{Block: true, Reason: strings.Join(lines, "\n")}, nil
}

func waitsForBuilder(s states.Status) bool { return s == states.Approved || s == states.Stale }
