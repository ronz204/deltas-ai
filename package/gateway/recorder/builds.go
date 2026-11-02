package recorder

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/diffing"
)

func (l *Ledger) buildFile(slice string) string {
	return filepath.Join(strings.TrimSuffix(l.file, ".json")+".builds", slice+".json")
}

func (l *Ledger) Snapshot(ctx context.Context, slice string) error {
	if l.file == "" {
		return errors.New("there is no cache directory to keep the snapshot in")
	}
	files, err := l.tree.Files(ctx, everything)
	if err != nil {
		return err
	}
	return save(l.buildFile(slice), files)
}

func (l *Ledger) Compare(ctx context.Context, slice string) (changed, gone []string, known bool, err error) {
	if l.file == "" {
		return nil, nil, false, nil
	}
	raw, err := os.ReadFile(l.buildFile(slice))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, false, err
	}
	var before map[string]string
	if json.Unmarshal(raw, &before) != nil {
		return nil, nil, false, nil //nolint:nilerr // a damaged snapshot is the same as none: the footprint is then left as it was
	}
	now, err := l.tree.Files(ctx, everything)
	if err != nil {
		return nil, nil, true, err
	}
	for _, name := range diffing.Changed(before, now) {
		if _, present := now[name]; present {
			changed = append(changed, name)
		} else {
			gone = append(gone, name)
		}
	}
	return changed, gone, true, nil
}

func (l *Ledger) Drop(_ context.Context, slice string) error {
	if l.file == "" {
		return nil
	}
	if err := os.Remove(l.buildFile(slice)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}
