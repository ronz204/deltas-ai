package deployer

import (
	"bytes"
	"context"
	"errors"
	"io/fs"

	"github.com/ronz204/deltas-ai/package/domain/manifest"
)

func (i *Installer) managed() ([]Asset, error) {
	assets, err := i.assets()
	if err != nil {
		return nil, err
	}
	var out []Asset
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
	return true, i.files.WriteFile(ctx, manifest.Name, want)
}

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
