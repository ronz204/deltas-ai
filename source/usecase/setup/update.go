package setup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/ronz204/deltas-ai/source/domain/manifest"
)

// managed lists the files delta keeps up to date: the git hook and the managed assets of the harness.
func (i *Installer) managed() ([]Asset, error) {
	hook, err := i.harness.GitHook(hookName)
	if err != nil {
		return nil, err
	}
	assets, err := i.harness.Assets(i.files.Name())
	if err != nil {
		return nil, err
	}
	out := []Asset{{Path: hookPath, Content: hook, Mode: 0o755, Managed: true}}
	for _, a := range assets {
		if a.Managed {
			out = append(out, a)
		}
	}
	return out, nil
}

func (i *Installer) readManifest(ctx context.Context) (manifest.Manifest, error) {
	text, err := i.files.ReadFile(ctx, manifest.Name)
	if errors.Is(err, fs.ErrNotExist) {
		return manifest.Parse(""), nil
	}
	if err != nil {
		return manifest.Manifest{}, err
	}
	return manifest.Parse(string(text)), nil
}

// writeManifest writes m unless the file already says the same, or there is nothing to say and no file.
func (i *Installer) writeManifest(ctx context.Context, m manifest.Manifest) (written bool, err error) {
	want := []byte(m.Format())
	have, err := i.files.ReadFile(ctx, manifest.Name)
	switch {
	case err == nil && bytes.Equal(have, want):
		return false, nil
	case err != nil && !errors.Is(err, fs.ErrNotExist):
		return false, err
	case err != nil && len(m.Files) == 0:
		return false, nil
	}
	return true, i.files.WriteFile(ctx, manifest.Name, want, 0o644)
}

// record notes the hash of the managed files that are in the project exactly as the harness has them. A file
// that differs is the project's to keep, and is not recorded.
func (i *Installer) record(ctx context.Context) ([]Step, error) {
	managed, err := i.managed()
	if err != nil {
		return nil, err
	}
	m, err := i.readManifest(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range managed {
		have, err := i.files.ReadFile(ctx, a.Path)
		switch {
		case errors.Is(err, fs.ErrNotExist):
		case err != nil:
			return nil, err
		case manifest.Hash(have) == manifest.Hash(a.Content):
			m.Files[a.Path] = manifest.Hash(a.Content)
		}
	}
	written, err := i.writeManifest(ctx, m)
	if err != nil || !written {
		return nil, err
	}
	return []Step{{true, "recorded the installed harness files in " + manifest.Name}}, nil
}

// plan is what update would do with each managed file, decided from three hashes: the one delta recorded
// when it installed the file, the file as it is now, and the file as this program ships it.
type plan struct {
	install []Asset  // missing from the project
	refresh []Asset  // untouched since delta installed them, and the harness has changed
	current int      // already as the harness has them
	edited  []string // changed by the project, or of unknown origin: left as they are
	remove  []string // no longer in the harness and untouched since delta installed them
	orphans []string // no longer in the harness but edited by the project: left, and forgotten
	known   manifest.Manifest
	record  manifest.Manifest // what the manifest says once the plan is done
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
		have, err := i.files.ReadFile(ctx, a.Path)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			p.install = append(p.install, a)
			p.record.Files[a.Path] = manifest.Hash(a.Content)
		case err != nil:
			return plan{}, err
		case manifest.Hash(have) == manifest.Hash(a.Content):
			p.current++
			p.record.Files[a.Path] = manifest.Hash(a.Content)
		case wasInstalled && manifest.Hash(have) == recorded:
			p.refresh = append(p.refresh, a)
			p.record.Files[a.Path] = manifest.Hash(a.Content)
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

// Pending lists, one line each, what update would change: the files it would refresh or remove and, in a
// project that already has a manifest, install. It writes nothing.
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

// Update brings the managed files up to date with this program's harness: it installs the ones that are
// missing, replaces those the project has not touched, removes those the harness no longer ships, and leaves
// alone every file the project changed. Settings are merged as init does. What it replaces or removes is a
// file the project did not edit, so git has it.
func (i *Installer) Update(ctx context.Context) ([]Step, error) {
	p, err := i.plan(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range append(append([]Asset(nil), p.install...), p.refresh...) {
		if err := i.files.WriteFile(ctx, a.Path, a.Content, a.Mode); err != nil {
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
