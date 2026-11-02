package vaults

import (
	"os"
	"path/filepath"

	"github.com/ronz204/deltas-ai/package/domain/linking"
)

const (
	dirName  = "deltas"
	specsDir = "specs"
	suffix   = ".spec.md"
)

var skipDirs = []string{".git", "node_modules", "vendor"}

type Vault struct {
	root   string
	isTest func(string) bool
	skip   map[string]bool
}

func New(root string) *Vault {
	return &Vault{root: root, isTest: linking.IsTestFile, skip: skipping(nil)}
}

func (s *Vault) Matching(patterns []string) *Vault {
	return &Vault{root: s.root, isTest: linking.TestMatcher(patterns), skip: s.skip}
}

func (s *Vault) Skipping(dirs []string) *Vault {
	return &Vault{root: s.root, isTest: s.isTest, skip: skipping(dirs)}
}

func skipping(dirs []string) map[string]bool {
	if len(dirs) == 0 {
		dirs = skipDirs
	}
	out := map[string]bool{dirName: true}
	for _, d := range dirs {
		out[d] = true
	}
	return out
}

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

func (s *Vault) Name() string { return filepath.Base(s.root) }
