// Package files reads and writes a project's knowledge base on disk.
package files

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ronz204/deltas-ai/internal/domain/lint"
	"github.com/ronz204/deltas-ai/internal/domain/spec"
)

const (
	dirName = "deltas"
	suffix  = ".spec.md"
)

// skipDirs are never searched for test files: they hold other people's code or the contracts.
var skipDirs = map[string]bool{".git": true, "node_modules": true, "vendor": true, dirName: true}

// Store is the knowledge base of the project rooted at one directory.
//
// shortcut: contracts live in <root>/deltas only; a repo split into services keeps one deltas/
// per service and is not read yet.
type Store struct{ root string }

func New(root string) *Store { return &Store{root: root} }

// FindRoot returns the nearest directory at or above start that holds a .git, or start itself
// when there is none.
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

// Path is where a slice's spec is shown to be, relative to the root and with forward slashes.
func (s *Store) Path(slice string) string { return dirName + "/" + slice + suffix }

// file is the real location of a slice's spec. It refuses a name that could point outside deltas/.
func (s *Store) file(slice string) (string, error) {
	if !spec.ValidSlice(slice) {
		return "", fmt.Errorf("%q is not a slice name; use lowercase words joined by hyphens", slice)
	}
	return filepath.Join(s.root, dirName, slice+suffix), nil
}

// List returns the names of the slices that have a spec, sorted.
func (s *Store) List(context.Context) ([]string, error) {
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

func (s *Store) Load(_ context.Context, slice string) ([]byte, error) {
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

// Create writes a new spec and fails if one already exists.
//
// shortcut: it checks and then writes, so two processes creating the same slice at once could
// both pass; the CLI is run by one person at a time.
func (s *Store) Create(_ context.Context, slice string, content []byte) error {
	p, err := s.file(slice)
	if err != nil {
		return err
	}
	if _, err := os.Stat(p); err == nil {
		return fmt.Errorf("%s already exists", s.Path(slice))
	}
	return writeAtomic(p, content)
}

// Save writes a spec, replacing the one that is there. The old file is replaced whole or not at
// all, so a spec is never left half written.
func (s *Store) Save(_ context.Context, slice string, content []byte) error {
	p, err := s.file(slice)
	if err != nil {
		return err
	}
	return writeAtomic(p, content)
}

// writeAtomic writes data to a temporary file beside path and renames it over path, so readers
// see the old content or the new, never half of it.
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

// TestFiles reads every test file under the root, in path order.
//
// shortcut: every test file is held in memory at once; for a very large repo, read only the
// files under the paths the claims name.
func (s *Store) TestFiles(ctx context.Context) ([]lint.File, error) {
	var out []lint.File
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
		if !lint.IsTestFile(rel) {
			return nil
		}
		text, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		out = append(out, lint.File{Path: rel, Text: string(text)})
		return nil
	})
	return out, err
}
