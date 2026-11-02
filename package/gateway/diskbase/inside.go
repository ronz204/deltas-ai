package diskbase

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/budget"
)

func (s *Vault) inside(rel string) (string, error) {
	local := filepath.FromSlash(rel)
	if strings.Contains(rel, `\`) || !filepath.IsLocal(local) {
		return "", fmt.Errorf("%q is not a path inside the project", rel)
	}
	return filepath.Join(s.root, local), nil
}

func (s *Vault) ReadFile(_ context.Context, rel string) ([]byte, error) {
	p, err := s.inside(rel)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(p)
}

func (s *Vault) WriteFile(_ context.Context, rel string, content []byte) error {
	p, err := s.inside(rel)
	if err != nil {
		return err
	}
	return writeAtomic(p, content)
}

func (s *Vault) RemoveFile(_ context.Context, rel string) error {
	p, err := s.inside(rel)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	for dir := filepath.Dir(p); filepath.Dir(dir) != s.root && dir != s.root; dir = filepath.Dir(dir) {
		if os.Remove(dir) != nil {
			break
		}
	}
	return nil //nolint:nilerr // pruning empty folders is best effort: stop at the first one that stays
}

func (s *Vault) Glob(_ context.Context, patterns []string) ([]budget.File, error) {
	root := os.DirFS(s.root)
	var out []budget.File
	for _, pattern := range patterns {
		names, err := fs.Glob(root, pattern)
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			if info, err := fs.Stat(root, name); err != nil || info.IsDir() {
				continue
			}
			text, err := fs.ReadFile(root, name)
			if err != nil {
				return nil, err
			}
			out = append(out, budget.File{Path: name, Text: string(text)})
		}
	}
	return out, nil
}

func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".delta-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	_, werr := tmp.Write(data)
	if err := errors.Join(werr, tmp.Close(), os.Chmod(tmp.Name(), 0o644)); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
