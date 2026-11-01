// Package author is the only write path of the knowledge base: everything is validated first.
package author

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/ronz204/deltas-ai/source/domain/assertion"
	"github.com/ronz204/deltas-ai/source/domain/findings"
	"github.com/ronz204/deltas-ai/source/domain/specfile"
)

type Store interface {
	Load(ctx context.Context, slice string) ([]byte, error)
	Create(ctx context.Context, slice string, content []byte) error
	Save(ctx context.Context, slice string, content []byte) error
	Path(slice string) string
}

type Templates interface {
	Spec() ([]byte, error)
}

type Author struct {
	store Store
	tpl   Templates
}

func New(store Store, tpl Templates) *Author { return &Author{store: store, tpl: tpl} }

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

// Write saves content only if it is a valid spec; otherwise nothing is written and the problems come back.
func (a *Author) Write(ctx context.Context, slice string, content []byte) ([]findings.Diagnostic, error) {
	if _, ds := specfile.ParseSlice(a.store.Path(slice), slice, content); len(ds) > 0 {
		return ds, nil
	}
	return nil, a.store.Save(ctx, slice, content)
}

func (a *Author) Link(ctx context.Context, slice, claimID, command string) error {
	id, err := assertion.ParseID(claimID)
	if err != nil {
		return err
	}
	tag, err := assertion.ParseTag("test: " + strings.TrimSpace(command)) // it also rejects an empty command
	if err != nil {
		return err
	}
	src, err := a.store.Load(ctx, slice)
	if err != nil {
		return err
	}
	edited, err := specfile.SetTag(src, id, tag)
	if err != nil {
		return err
	}
	return a.store.Save(ctx, slice, edited)
}
