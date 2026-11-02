// Package status summarises where every slice stands.
package status

import (
	"context"
	"fmt"
	"slices"

	"github.com/ronz204/deltas-ai/package/domain/budget"
	"github.com/ronz204/deltas-ai/package/domain/claimset"
	"github.com/ronz204/deltas-ai/package/domain/guides"
	"github.com/ronz204/deltas-ai/package/domain/linking"
	"github.com/ronz204/deltas-ai/package/domain/specfile"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

type Specs interface {
	List(ctx context.Context) ([]string, error)
	Load(ctx context.Context, slice string) ([]byte, error)
	Path(slice string) string
}

// Baselines knows what a slice was verified against. Drift lists the files of its scope that changed since;
// known is false when it was never verified here, and then there is nothing to compare.
type Baselines interface {
	Drift(ctx context.Context, slice string, scope linking.Scope) (changed []string, known bool, err error)
}

// Saver writes a spec back; the reporter uses it only to mark a slice stale.
type Saver interface {
	Save(ctx context.Context, slice string, content []byte) error
}

// Files reads the project files that match patterns, where * stands for part of one folder or file name.
type Files interface {
	Glob(ctx context.Context, patterns []string) ([]budget.File, error)
}

type Reporter struct {
	specs     Specs
	baselines Baselines
	saver     Saver
	files     Files
}

// New builds the reporter.
func New(specs Specs) *Reporter {
	return &Reporter{specs: specs}
}

// WithDrift makes the reporter notice slices whose files changed after they were verified, and mark them
// stale. Without it nothing is ever stale. It returns the reporter.
func (r *Reporter) WithDrift(b Baselines, s Saver) *Reporter {
	r.baselines, r.saver = b, s
	return r
}

// WithDocs makes the reporter able to say how many of the docs for people are written. It returns the reporter.
func (r *Reporter) WithDocs(f Files) *Reporter {
	r.files = f
	return r
}

// Row counts a slice's claims by tag; Issues counts the problems found, and the counts cover only clean claims.
type Row struct {
	Slice  string
	Status states.Status // empty when the frontmatter could not be read
	Claims int
	Test   int
	Manual int
	None   int
	Issues int
	Scope  linking.Scope // the part of the project the slice depends on

	Changed []string // for a stale slice, the files that changed since it was verified
}

// Report lists every slice. When drift is on, it first marks as stale the verified slices whose files
// changed; if that cannot be told (no git, say) the slices are reported as the specs say.
func (r *Reporter) Report(ctx context.Context) ([]Row, error) {
	drifted, _ := r.Stale(ctx)
	changed := map[string][]string{}
	for _, d := range drifted {
		changed[d.Slice] = d.Files
	}
	names, err := r.specs.List(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]Row, 0, len(names))
	for _, name := range names {
		src, err := r.specs.Load(ctx, name)
		if err != nil {
			return nil, err
		}
		file := r.specs.Path(name)
		sp, ds := specfile.ParseSlice(file, name, src)
		row := Row{Slice: name, Status: sp.Status, Claims: len(sp.Claims), Issues: len(ds), Scope: linking.NewScope(file), Changed: changed[name]}
		for _, c := range sp.Claims {
			switch c.Tag.Method {
			case claimset.Test:
				row.Test++
			case claimset.Manual:
				row.Manual++
			default:
				row.None++
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// NotReady is why a slice cannot be built yet.
type NotReady struct{ Reason string }

func (e *NotReady) Error() string { return e.Reason }

// Ready returns a *NotReady unless the slice has a clean spec and is approved or stale: the contract Builder builds from.
func (r *Reporter) Ready(ctx context.Context, slice string) error {
	src, err := r.specs.Load(ctx, slice)
	if err != nil {
		return err
	}
	sp, ds := specfile.ParseSlice(r.specs.Path(slice), slice, src)
	switch {
	case len(ds) > 0:
		return &NotReady{fmt.Sprintf("the spec of %s has %d problems; run delta spec lint %s", slice, len(ds), slice)}
	case sp.Status == states.Draft:
		return &NotReady{fmt.Sprintf("%s is a draft; a person approves it first with delta approve %s", slice, slice)}
	case sp.Status != states.Approved && sp.Status != states.Stale:
		return &NotReady{fmt.Sprintf("%s is %s, not approved; run delta check %s to see where it stands", slice, sp.Status, slice)}
	}
	return nil
}

// Drift is a slice whose files changed after it was verified. Marked is true when this call is what made it stale.
type Drift struct {
	Slice  string
	Files  []string
	Marked bool
}

// Stale compares each verified or stale slice with the files it was verified against. A verified slice that
// drifted becomes stale in its spec; one already stale stays so. Slices that are not drifting are not listed.
// The comparison is by file: any change to a covered file counts, whether or not it alters behaviour (direction.md, §10).
func (r *Reporter) Stale(ctx context.Context) ([]Drift, error) {
	if r.baselines == nil || r.saver == nil {
		return nil, nil
	}
	names, err := r.specs.List(ctx)
	if err != nil {
		return nil, err
	}
	var out []Drift
	for _, name := range names {
		src, err := r.specs.Load(ctx, name)
		if err != nil {
			return out, err
		}
		file := r.specs.Path(name)
		sp, ds := specfile.ParseSlice(file, name, src)
		if len(ds) > 0 || (sp.Status != states.Verified && sp.Status != states.Stale) {
			continue
		}
		files, known, err := r.baselines.Drift(ctx, name, linking.NewScope(file))
		if err != nil {
			return out, err
		}
		if sp.Status == states.Stale { // marking it stale rewrote its spec, which is no news
			files = slices.DeleteFunc(files, func(f string) bool { return f == file })
		}
		if !known || len(files) == 0 {
			continue
		}
		d := Drift{Slice: name, Files: files}
		if sp.Status == states.Verified {
			edited, err := specfile.SetStatus(src, states.Stale)
			if err != nil {
				return out, err
			}
			if err := r.saver.Save(ctx, name, edited); err != nil {
				return out, err
			}
			d.Marked = true
		}
		out = append(out, d)
	}
	return out, nil
}

// DocsRow says how many of the docs for people are written, and which one to write next.
type DocsRow struct {
	Total  int
	Filled int
	Next   string // the first doc, in the order guides.Order gives, that is not written yet; empty when all are
}

// Docs counts the templates delta installed under deltas/docs/ and how many of them were written. It reports
// nothing (a zero row) for a project with no docs or when the reporter cannot read files.
func (r *Reporter) Docs(ctx context.Context) (DocsRow, error) {
	if r.files == nil {
		return DocsRow{}, nil
	}
	found, err := r.files.Glob(ctx, []string{guides.Dir + "/*.md"})
	if err != nil {
		return DocsRow{}, err
	}
	var row DocsRow
	open := map[string]bool{}
	for _, f := range found {
		name, ok := guides.Name(f.Path)
		if !ok {
			continue
		}
		row.Total++
		if guides.Filled(f.Text) {
			row.Filled++
		} else {
			open[name] = true
		}
	}
	for _, name := range guides.Order {
		if open[name] {
			row.Next = name
			break
		}
	}
	if row.Next == "" { // a doc outside the usual four
		for name := range open {
			if row.Next == "" || name < row.Next {
				row.Next = name
			}
		}
	}
	return row, nil
}
