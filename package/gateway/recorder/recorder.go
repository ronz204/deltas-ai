package recorder

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ronz204/deltas-ai/package/domain/diffing"
	"github.com/ronz204/deltas-ai/package/domain/judgment"
	"github.com/ronz204/deltas-ai/package/domain/linking"
	"github.com/ronz204/deltas-ai/source/usecase/verify"
)

type Tree interface {
	Tree(ctx context.Context, scope linking.Scope) (hash string, dirty bool, err error)
	Files(ctx context.Context, scope linking.Scope) (map[string]string, error)
}

type Ledger struct {
	file string
	tree Tree
}

type entry struct {
	Tree     string            `json:"tree"`
	Violated bool              `json:"violated"`
	Files    map[string]string `json:"files,omitempty"`
}

func New(cache, root string, tree Tree) *Ledger {
	l := &Ledger{tree: tree}
	if cache != "" {
		sum := sha256.Sum256([]byte(root))
		l.file = filepath.Join(cache, "delta", hex.EncodeToString(sum[:8])+".json")
	}
	return l
}

func (l *Ledger) Record(ctx context.Context, seen []verify.Seen) error {
	if l.file == "" {
		return errors.New("there is no cache directory to keep the record in")
	}
	marks := l.read()
	for _, s := range seen {
		tree, _, err := l.tree.Tree(ctx, s.Scope)
		if err != nil {
			return err
		}
		e := marks[s.Slice]
		e.Tree, e.Violated = tree, s.Violated
		if s.Verified {
			if e.Files, err = l.tree.Files(ctx, s.Scope); err != nil {
				return err
			}
		}
		marks[s.Slice] = e
	}
	return l.write(marks)
}

func (l *Ledger) Drift(ctx context.Context, slice string, scope linking.Scope) (changed []string, known bool, err error) {
	files := l.read()[slice].Files
	if files == nil {
		return nil, false, nil
	}
	now, err := l.tree.Files(ctx, scope)
	if err != nil {
		return nil, true, err
	}
	return diffing.Changed(files, now), true, nil
}

func (l *Ledger) Standings(ctx context.Context, slices map[string]linking.Scope) (map[string]judgment.Standing, error) {
	marks := l.read()
	by := make(map[string]judgment.Standing, len(slices))
	for slice, scope := range slices {
		tree, dirty, err := l.tree.Tree(ctx, scope)
		if err != nil {
			return nil, err
		}
		if !dirty {
			by[slice] = judgment.Passing
			continue
		}
		e := marks[slice]
		by[slice] = judgment.Mark{Tree: e.Tree, Violated: e.Violated}.At(tree)
	}
	return by, nil
}

func (l *Ledger) read() map[string]entry {
	var marks map[string]entry
	if raw, err := os.ReadFile(l.file); err == nil && json.Unmarshal(raw, &marks) != nil {
		marks = nil
	}
	if marks == nil {
		marks = map[string]entry{}
	}
	return marks
}

func (l *Ledger) write(marks map[string]entry) error {
	raw, err := json.MarshalIndent(marks, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(l.file)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".record-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, werr := tmp.Write(raw)
	if err := errors.Join(werr, tmp.Close()); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), l.file); err != nil {
		return fmt.Errorf("save the record: %w", err)
	}
	return nil
}
