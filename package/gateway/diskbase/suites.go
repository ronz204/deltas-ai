package diskbase

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ronz204/deltas-ai/package/domain/linking"
)

func (s *Vault) IsTest(rel string) bool { return s.isTest(rel) }

func (s *Vault) TestFiles(ctx context.Context) ([]linking.File, error) {
	var out []linking.File
	err := filepath.WalkDir(s.root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != s.root && s.skip[d.Name()] {
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
		out = append(out, linking.File{Path: rel, Text: string(text)})
		return nil
	})
	return out, err
}
