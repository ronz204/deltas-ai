// Package author is the only way the knowledge base is written: everything is validated before
// it reaches the disk.
package author

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/ronz204/deltas-ai/internal/domain/claim"
	"github.com/ronz204/deltas-ai/internal/domain/diag"
	"github.com/ronz204/deltas-ai/internal/domain/spec"
)

// Store is where specs are kept.
type Store interface {
	Load(ctx context.Context, slice string) ([]byte, error)
	Create(ctx context.Context, slice string, content []byte) error
	Save(ctx context.Context, slice string, content []byte) error
	Path(slice string) string
}

// Templates gives the text a new spec starts from.
type Templates interface {
	Spec() ([]byte, error)
}

type Author struct {
	store Store
	tpl   Templates
}

func New(store Store, tpl Templates) *Author { return &Author{store: store, tpl: tpl} }

// Create makes deltas/<slice>.spec.md from the template, in the draft state. It returns the path.
func (a *Author) Create(ctx context.Context, slice string) (string, error) {
	raw, err := a.tpl.Spec()
	if err != nil {
		return "", err
	}
	tpl, err := template.New("spec").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("the spec template is broken: %w", err)
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, struct{ Slice string }{slice}); err != nil {
		return "", err
	}
	if err := a.store.Create(ctx, slice, out.Bytes()); err != nil {
		return "", err
	}
	return a.store.Path(slice), nil
}

// Write validates content as the spec of slice and saves it. If it has any problem nothing is
// written and the problems come back; the error is for failures of the store itself.
func (a *Author) Write(ctx context.Context, slice string, content []byte) ([]diag.Diagnostic, error) {
	if _, ds := spec.ParseSlice(a.store.Path(slice), slice, content); len(ds) > 0 {
		return ds, nil
	}
	return nil, a.store.Save(ctx, slice, content)
}

// Link makes the claim verified by command: its tag becomes test: <command>. Only that line
// of the spec changes.
func (a *Author) Link(ctx context.Context, slice, claimID, command string) error {
	id, err := claim.ParseID(claimID)
	if err != nil {
		return err
	}
	tag, err := claim.ParseTag("test: " + strings.TrimSpace(command)) // it also rejects an empty command
	if err != nil {
		return err
	}
	src, err := a.store.Load(ctx, slice)
	if err != nil {
		return err
	}
	edited, err := spec.SetTag(src, id, tag)
	if err != nil {
		return err
	}
	return a.store.Save(ctx, slice, edited)
}
