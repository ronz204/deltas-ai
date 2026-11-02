package recorder

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ronz204/deltas-ai/package/domain/diffing"
)

const sessionTTL = 14 * 24 * time.Hour

func (l *Ledger) sessionFile(session string) string {
	sum := sha256.Sum256([]byte(session))
	return filepath.Join(strings.TrimSuffix(l.file, ".json")+".sessions", hex.EncodeToString(sum[:8])+".json")
}

func (l *Ledger) Begin(ctx context.Context, session string) error {
	if l.file == "" {
		return errors.New("there is no cache directory to keep the session in")
	}
	file := l.sessionFile(session)
	if _, err := os.Stat(file); err == nil {
		return nil
	}
	b, err := l.tree.Baseline(ctx)
	if err != nil {
		return err
	}
	if err := save(file, b); err != nil {
		return err
	}
	prune(filepath.Dir(file))
	return nil
}

func (l *Ledger) Touched(ctx context.Context, session string) (files []string, known bool, err error) {
	if l.file == "" {
		return nil, false, nil
	}
	raw, err := os.ReadFile(l.sessionFile(session))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var b diffing.Baseline
	if json.Unmarshal(raw, &b) != nil {
		return nil, false, nil //nolint:nilerr // a damaged baseline is the same as none: the Stop hook then lets Claude finish
	}
	files, err = l.tree.Since(ctx, b)
	return files, true, err
}

func prune(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if info, err := e.Info(); err == nil && time.Since(info.ModTime()) > sessionTTL {
			os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}
