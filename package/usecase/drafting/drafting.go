package drafting

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"text/template"

	"github.com/ronz204/deltas-ai/package/domain/claimset"
	"github.com/ronz204/deltas-ai/package/domain/english"
	"github.com/ronz204/deltas-ai/package/domain/findings"
	"github.com/ronz204/deltas-ai/package/domain/specfile"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

const (
	CodeNewStatus  = "AUTHOR-NEW-STATUS"
	CodeSetStatus  = "AUTHOR-SET-STATUS"
	CodeReapproval = "AUTHOR-REAPPROVE"
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

func (a *Author) Write(ctx context.Context, slice string, content []byte) ([]findings.Diagnostic, error) {
	file := a.store.Path(slice)
	sp, ds := specfile.ParseSlice(file, slice, content)
	if len(ds) > 0 {
		return ds, nil
	}
	content, ds, err := a.settle(ctx, file, slice, sp.Status, content)
	if err != nil || hasError(ds) {
		return ds, err
	}
	if english.NotEnglish(string(content)) {
		ds = append(ds, findings.Diagnostic{File: file, Code: english.Code, Warn: true,
			Message: "the spec is not in English; the knowledge base is written in English whatever language the user speaks: translate it and write it again"})
	}
	return ds, a.store.Save(ctx, slice, content)
}

func (a *Author) settle(ctx context.Context, file, slice string, status states.Status, content []byte) ([]byte, []findings.Diagnostic, error) {
	prev, err := a.store.Load(ctx, slice)
	if err != nil {
		if status != states.Draft {
			return nil, []findings.Diagnostic{{File: file, Line: 1, Code: CodeNewStatus,
				Message: fmt.Sprintf("a new spec is written as a draft, not %s; delta approve moves it forward", status)}}, nil
		}
		return content, nil, nil
	}
	old, _ := specfile.ParseSlice(file, slice, prev)
	if status != states.Draft && status != old.Status {
		return nil, []findings.Diagnostic{{File: file, Line: 1, Code: CodeSetStatus,
			Message: fmt.Sprintf("the spec is %q and cannot be written as %q; only the delta commands change the status", old.Status, status)}}, nil
	}
	if old.Status == "" || old.Status == states.Draft || sameContract(prev, content) {
		return content, nil, nil
	}
	redrafted, err := specfile.SetStatus(content, states.Draft)
	if err != nil {
		return nil, nil, err
	}
	return redrafted, []findings.Diagnostic{{File: file, Code: CodeReapproval, Warn: true,
		Message: fmt.Sprintf("the contract changed, so %s is a draft again; run delta approve %s once it is right", slice, slice)}}, nil
}

func sameContract(a, b []byte) bool {
	x, errA := specfile.SetStatus(a, states.Draft)
	y, errB := specfile.SetStatus(b, states.Draft)
	return errA == nil && errB == nil && bytes.Equal(x, y)
}

func hasError(ds []findings.Diagnostic) bool {
	return slices.ContainsFunc(ds, func(d findings.Diagnostic) bool { return !d.Warn })
}

func (a *Author) Approve(ctx context.Context, slice string) ([]findings.Diagnostic, error) {
	return a.move(ctx, slice, states.Approved)
}

func (a *Author) Built(ctx context.Context, slice string) ([]findings.Diagnostic, error) {
	return a.move(ctx, slice, states.Built)
}

func (a *Author) move(ctx context.Context, slice string, to states.Status) ([]findings.Diagnostic, error) {
	src, err := a.store.Load(ctx, slice)
	if err != nil {
		return nil, err
	}
	sp, ds := specfile.ParseSlice(a.store.Path(slice), slice, src)
	if len(ds) > 0 {
		return ds, nil
	}
	if !states.Can(sp.Status, to) {
		return nil, fmt.Errorf("%s is %s and cannot become %s", slice, sp.Status, to)
	}
	if to == states.Approved && len(sp.Claims) == 0 {
		return nil, fmt.Errorf("%s has no claims, so there is nothing to approve", slice)
	}
	edited, err := specfile.SetStatus(src, to)
	if err != nil {
		return nil, err
	}
	return nil, a.store.Save(ctx, slice, edited)
}

func (a *Author) Link(ctx context.Context, slice, claimID, command string) error {
	id, err := claimset.ParseID(claimID)
	if err != nil {
		return err
	}
	tag, err := claimset.ParseTag("test: " + strings.TrimSpace(command))
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
