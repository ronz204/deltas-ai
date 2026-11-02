package diskbase

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func (s *Vault) printsFile(slice string) (string, error) {
	spec, err := s.file(slice)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(spec, suffix) + prints, nil
}

func (s *Vault) Footprint(_ context.Context, slice string) ([]string, error) {
	p, err := s.printsFile(slice)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(string(raw), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, filepath.ToSlash(line))
		}
	}
	return files, nil
}

func (s *Vault) SaveFootprint(_ context.Context, slice string, files []string) error {
	p, err := s.printsFile(slice)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	return writeAtomic(p, []byte(strings.Join(files, "\n")+"\n"))
}
