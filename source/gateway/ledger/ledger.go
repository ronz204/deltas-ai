// Package ledger remembers, per project, which tree each slice was last checked against. It lives in the
// user's cache directory, so it never touches the project or its .git; losing it only makes delta ask for a new check.
package ledger

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

// Tree fingerprints the part of the working tree a scope covers; dirty is false when it has no changes.
// Files fingerprints every file the scope covers, one hash each, whatever HEAD says.
type Tree interface {
	Tree(ctx context.Context, scope linking.Scope) (hash string, dirty bool, err error)
	Files(ctx context.Context, scope linking.Scope) (map[string]string, error)
}

type Ledger struct {
	file string
	tree Tree
}

// New keeps the project at root in a file inside cache. An empty cache makes every call fail.
func New(cache, root string, tree Tree) *Ledger {
	l := &Ledger{tree: tree}
	if cache != "" {
		sum := sha256.Sum256([]byte(root))
		l.file = filepath.Join(cache, "delta", hex.EncodeToString(sum[:8])+".json")
	}
	return l
}

// entry is what the last check left for a slice. Files is set only when that check ended with the slice
// verified, and stays until another one does: it is what the slice is compared with to find drift.
// shortcut: a slice that depends on every code file keeps a hash per file; upgrade by sharing one table between slices.
type entry struct {
	Tree     string            `json:"tree"`
	Violated bool              `json:"violated"`
	Files    map[string]string `json:"files,omitempty"`
}

// Record marks each slice as checked against its part of the tree as it is now.
func (l *Ledger) Record(ctx context.Context, seen []verify.Seen) error {
	if l.file == "" {
		return errors.New("there is no cache directory to keep the record in")
	}
	marks, _ := l.read() // a missing or damaged record is replaced
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

// Drift lists the files of the scope that differ from the ones the slice was last verified against. known is
// false when it was never verified here (a fresh clone, or a lost record): then there is nothing to compare.
func (l *Ledger) Drift(ctx context.Context, slice string, scope linking.Scope) (changed []string, known bool, err error) {
	marks, _ := l.read()
	e := marks[slice]
	if e.Files == nil {
		return nil, false, nil
	}
	now, err := l.tree.Files(ctx, scope)
	if err != nil {
		return nil, true, err
	}
	return diffing.Changed(e.Files, now), true, nil
}

// Standings says what the last check of each slice proves about its part of the tree as it is now. A slice
// whose part has no changes has nothing to prove and stands.
func (l *Ledger) Standings(ctx context.Context, slices map[string]linking.Scope) (map[string]judgment.Standing, error) {
	marks, _ := l.read()
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

func (l *Ledger) read() (map[string]entry, error) {
	marks := map[string]entry{}
	if l.file == "" {
		return marks, errors.New("no cache directory")
	}
	raw, err := os.ReadFile(l.file)
	if err != nil {
		return marks, err
	}
	if err := json.Unmarshal(raw, &marks); err != nil || marks == nil {
		return map[string]entry{}, err
	}
	return marks, nil
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
	defer os.Remove(tmp.Name()) // after the rename the name is gone and this does nothing
	_, werr := tmp.Write(raw)
	if err := errors.Join(werr, tmp.Close()); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), l.file); err != nil {
		return fmt.Errorf("save the record: %w", err)
	}
	return nil
}
