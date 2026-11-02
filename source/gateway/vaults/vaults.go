// Package vaults reads and writes a project's knowledge base on disk.
package vaults

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ronz204/deltas-ai/source/domain/budget"
	"github.com/ronz204/deltas-ai/source/domain/linkcheck"
	"github.com/ronz204/deltas-ai/source/domain/specfile"
)

const (
	dirName  = "deltas"
	specsDir = "specs"
	suffix   = ".spec.md"
)

// skipDirs are never searched for test files: they hold other people's code or the contracts.
var skipDirs = map[string]bool{".git": true, "node_modules": true, "vendor": true, dirName: true}

// Contracts live in the project's deltas/specs/ folder, one <slice>.spec.md each.
type Vault struct {
	root   string
	isTest func(string) bool
}

func New(root string) *Vault { return &Vault{root: root, isTest: linkcheck.IsTestFile} }

// Matching is a vault that takes as test files those whose name matches one of the patterns.
func (s *Vault) Matching(patterns []string) *Vault {
	return &Vault{root: s.root, isTest: linkcheck.TestMatcher(patterns)}
}

func (s *Vault) IsTest(rel string) bool { return s.isTest(rel) }

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

func (s *Vault) Path(slice string) string { return path.Join(dirName, specsDir, slice+suffix) }

// Name is what the project is called: the name of its folder.
func (s *Vault) Name() string { return filepath.Base(s.root) }

// file is where a slice's spec lives; a name that is not a valid slice name has none.
func (s *Vault) file(slice string) (string, error) {
	if !specfile.ValidSlice(slice) {
		return "", fmt.Errorf("%q is not a slice name; use lowercase words joined by hyphens", slice)
	}
	return filepath.Join(s.root, dirName, specsDir, slice+suffix), nil
}

// List gives the name of every slice, sorted. A project with no specs/ folder has none.
func (s *Vault) List(context.Context) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(s.root, dirName, specsDir))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if name, ok := strings.CutSuffix(e.Name(), suffix); ok && !e.IsDir() {
			ids = append(ids, name)
		}
	}
	return ids, nil // ReadDir sorts by file name
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
	return writeAtomic(p, content, 0o644)
}

func (s *Vault) Save(_ context.Context, slice string, content []byte) error {
	p, err := s.file(slice)
	if err != nil {
		return err
	}
	return writeAtomic(p, content, 0o644)
}

// inside turns a project path written with forward slashes into a file path, and refuses one that
// leaves the project.
func (s *Vault) inside(rel string) (string, error) {
	local := filepath.FromSlash(rel)
	if strings.Contains(rel, `\`) || !filepath.IsLocal(local) {
		return "", fmt.Errorf("%q is not a path inside the project", rel)
	}
	return filepath.Join(s.root, local), nil
}

// ReadFile returns a project file; the error wraps fs.ErrNotExist when it is not there.
func (s *Vault) ReadFile(_ context.Context, rel string) ([]byte, error) {
	p, err := s.inside(rel)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(p)
}

func (s *Vault) WriteFile(_ context.Context, rel string, content []byte, perm fs.FileMode) error {
	p, err := s.inside(rel)
	if err != nil {
		return err
	}
	return writeAtomic(p, content, perm)
}

// RemoveFile deletes a project file, and then the folders it leaves empty, except the top-level ones.
func (s *Vault) RemoveFile(_ context.Context, rel string) error {
	p, err := s.inside(rel)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	for dir := filepath.Dir(p); filepath.Dir(dir) != s.root && dir != s.root; dir = filepath.Dir(dir) {
		if os.Remove(dir) != nil { // not empty, or not ours to remove
			break
		}
	}
	return nil
}

// Glob reads the project files whose path matches any pattern, where * stands for part of one folder
// or file name. Paths come back relative to the root and with forward slashes.
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

// writeAtomic renames a temporary file over path, so readers see the old content or the new, never half.
func writeAtomic(path string, data []byte, perm fs.FileMode) error {
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
	if err := os.Chmod(tmp.Name(), perm); err != nil {
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
		if !s.isTest(rel) {
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
