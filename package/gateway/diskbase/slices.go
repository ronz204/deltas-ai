package diskbase

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/specfile"
)

func (s *Vault) Path(slice string) string { return path.Join(dirName, specsDir, slice+suffix) }

func (s *Vault) file(slice string) (string, error) {
	if !specfile.ValidSlice(slice) {
		return "", fmt.Errorf("%q is not a slice name; use lowercase words joined by hyphens", slice)
	}
	return filepath.Join(s.root, dirName, specsDir, slice+suffix), nil
}

func (s *Vault) List(context.Context) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(s.root, dirName, specsDir))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var slices []string
	for _, e := range entries {
		if name, ok := strings.CutSuffix(e.Name(), suffix); ok && !e.IsDir() {
			slices = append(slices, name)
		}
	}
	return slices, nil
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
