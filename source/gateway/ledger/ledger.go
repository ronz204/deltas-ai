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

	"github.com/ronz204/deltas-ai/source/domain/judgment"
)

// Tree fingerprints the working tree; dirty is false when it has no changes.
type Tree interface {
	Tree(ctx context.Context) (hash string, dirty bool, err error)
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

type entry struct {
	Tree     string `json:"tree"`
	Violated bool   `json:"violated"`
}

// Record marks each slice as checked against the tree as it is now; the value says if a claim was violated.
func (l *Ledger) Record(ctx context.Context, violated map[string]bool) error {
	if l.file == "" {
		return errors.New("there is no cache directory to keep the record in")
	}
	tree, _, err := l.tree.Tree(ctx)
	if err != nil {
		return err
	}
	marks, _ := l.read() // a missing or damaged record is replaced
	for slice, v := range violated {
		marks[slice] = entry{Tree: tree, Violated: v}
	}
	return l.write(marks)
}

// Standings says what the last check of each slice proves about the tree as it is now. dirty is false when
// the tree has no changes, and then there is nothing to prove.
func (l *Ledger) Standings(ctx context.Context, slices []string) (dirty bool, by map[string]judgment.Standing, err error) {
	tree, dirty, err := l.tree.Tree(ctx)
	if err != nil || !dirty {
		return false, nil, err
	}
	marks, _ := l.read()
	by = make(map[string]judgment.Standing, len(slices))
	for _, s := range slices {
		e := marks[s]
		by[s] = judgment.Mark{Tree: e.Tree, Violated: e.Violated}.At(tree)
	}
	return true, by, nil
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
