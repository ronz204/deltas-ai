package deployer

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/manifest"
)

type plan struct {
	install []Asset
	refresh []Asset
	current int
	edited  []string
	remove  []string
	orphans []string
	known   manifest.Manifest
	record  manifest.Manifest
}

func (i *Installer) plan(ctx context.Context) (plan, error) {
	managed, err := i.managed()
	if err != nil {
		return plan{}, err
	}
	known, err := i.readManifest(ctx)
	if err != nil {
		return plan{}, err
	}
	p := plan{known: known, record: manifest.Manifest{Files: map[string]string{}}}
	shipped := map[string]bool{}
	for _, a := range managed {
		shipped[a.Path] = true
		recorded, wasInstalled := known.Files[a.Path]
		want := manifest.Hash(a.Content)
		have, err := i.files.ReadFile(ctx, a.Path)
		missing := errors.Is(err, fs.ErrNotExist)
		if err != nil && !missing {
			return plan{}, err
		}
		switch {
		case missing:
			p.install = append(p.install, a)
			p.record.Files[a.Path] = want
		case manifest.Hash(have) == want:
			p.current++
			p.record.Files[a.Path] = want
		case wasInstalled && manifest.Hash(have) == recorded:
			p.refresh = append(p.refresh, a)
			p.record.Files[a.Path] = want
		default:
			p.edited = append(p.edited, a.Path)
			if wasInstalled {
				p.record.Files[a.Path] = recorded
			}
		}
	}
	for path, recorded := range known.Files {
		if shipped[path] {
			continue
		}
		have, err := i.files.ReadFile(ctx, path)
		switch {
		case errors.Is(err, fs.ErrNotExist):
		case err != nil:
			return plan{}, err
		case manifest.Hash(have) == recorded:
			p.remove = append(p.remove, path)
		default:
			p.orphans = append(p.orphans, path)
		}
	}
	slices.Sort(p.remove)
	slices.Sort(p.orphans)
	return p, nil
}

func (i *Installer) Pending(ctx context.Context) ([]string, error) {
	p, err := i.plan(ctx)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, a := range p.refresh {
		out = append(out, "update "+a.Path)
	}
	if len(p.known.Files) > 0 {
		for _, a := range p.install {
			out = append(out, "install "+a.Path)
		}
	}
	for _, path := range p.remove {
		out = append(out, "remove "+path)
	}
	slices.Sort(out)
	return out, nil
}

func (i *Installer) Update(ctx context.Context) ([]Step, error) {
	p, err := i.plan(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range append(append([]Asset(nil), p.install...), p.refresh...) {
		if err := i.files.WriteFile(ctx, a.Path, a.Content); err != nil {
			return nil, err
		}
	}
	for _, path := range p.remove {
		if err := i.files.RemoveFile(ctx, path); err != nil {
			return nil, err
		}
	}
	settings, err := i.installSettings(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := i.writeManifest(ctx, p.record); err != nil {
		return nil, err
	}

	var steps []Step
	add := func(done bool, format string, args ...any) {
		steps = append(steps, Step{done, fmt.Sprintf(format, args...)})
	}
	if n := len(p.install); n > 0 {
		add(true, "installed %d new %s from the harness: %s", n, noun(n), strings.Join(paths(p.install), ", "))
	}
	if n := len(p.refresh); n > 0 {
		add(true, "updated %d %s to this program's harness: %s", n, noun(n), strings.Join(paths(p.refresh), ", "))
	}
	if n := len(p.remove); n > 0 {
		add(true, "removed %d %s the harness no longer ships: %s", n, noun(n), strings.Join(p.remove, ", "))
	}
	if len(p.edited) > 0 {
		add(false, "left as they are, since the project changed them or delta did not install them: %s", strings.Join(p.edited, ", "))
	}
	if len(p.orphans) > 0 {
		add(false, "no longer in the harness, but the project changed them: left as they are: %s", strings.Join(p.orphans, ", "))
	}
	if len(steps) == 0 {
		add(false, "the %d managed harness files are already up to date", p.current)
	}
	return append(steps, settings), nil
}

func paths(assets []Asset) []string {
	out := make([]string, len(assets))
	for i, a := range assets {
		out[i] = a.Path
	}
	return out
}
