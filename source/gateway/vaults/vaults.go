// Package vaults reads and writes a project's knowledge base on disk.
package vaults

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ronz204/deltas-ai/source/domain/linkcheck"
	"github.com/ronz204/deltas-ai/source/domain/specfile"
)

const (
	dirName = "deltas"
	suffix  = ".spec.md"
)

// skipDirs are never searched for test files: they hold other people's code or the contracts.
var skipDirs = map[string]bool{".git": true, "node_modules": true, "vendor": true, dirName: true}

// shortcut: contracts are read from <root>/deltas only; a repo split into services is not supported yet.
type Vault struct{ root string }

func New(root string) *Vault { return &Vault{root: root} }

func FindRoot(start string) string {
	for dir := start; ; {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return start
		}
		dir = parent
	}
}

func (s *Vault) Path(slice string) string { return dirName + "/" + slice + suffix }

// file refuses a name that could point outside deltas/.
func (s *Vault) file(slice string) (string, error) {
	if !specfile.ValidSlice(slice) {
		return "", fmt.Errorf("%q is not a slice name; use lowercase words joined by hyphens", slice)
	}
	return filepath.Join(s.root, dirName, slice+suffix), nil
}

func (s *Vault) List(context.Context) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(s.root, dirName))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries { // ReadDir already returns them sorted by name
		if name, ok := strings.CutSuffix(e.Name(), suffix); ok && !e.IsDir() {
			names = append(names, name)
		}
	}
	return names, nil
}

func (s *Vault) Load(_ context.Context, slice string) ([]byte, error) {
	p, err := s.file(slice)
	if err != nil {
		return nil, err
	}
	src, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("there is no spec for %q (expected %s)", slice, s.Path(slice))
	}
	return src, err
}

// shortcut: check-then-write, so two simultaneous creates could both pass; one person runs the CLI at a time.
func (s *Vault) Create(_ context.Context, slice string, content []byte) error {
	p, err := s.file(slice)
	if err != nil {
		return err
	}
	if _, err := os.Stat(p); err == nil {
		return fmt.Errorf("%s already exists", s.Path(slice))
	}
	return writeAtomic(p, content)
}

func (s *Vault) Save(_ context.Context, slice string, content []byte) error {
	p, err := s.file(slice)
	if err != nil {
		return err
	}
	return writeAtomic(p, content)
}

// writeAtomic renames a temporary file over path, so readers see the old content or the new, never half.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".delta-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // after the rename the name is gone and this does nothing

	_, err = tmp.Write(data)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// shortcut: every test file is held in memory; for a very large repo read only the files under the claims' paths.
func (s *Vault) TestFiles(ctx context.Context) ([]linkcheck.File, error) {
	var out []linkcheck.File
	err := filepath.WalkDir(s.root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != s.root && skipDirs[d.Name()] {
				return fs.SkipDir
			}
			return ctx.Err()
		}
		rel, err := filepath.Rel(s.root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !linkcheck.IsTestFile(rel) {
			return nil
		}
		text, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		out = append(out, linkcheck.File{Path: rel, Text: string(text)})
		return nil
	})
	return out, err
}
