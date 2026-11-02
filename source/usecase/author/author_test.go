package author

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/specfile"
)

var ctx = context.Background()

type memory struct {
	specs map[string]string
	fail  error
}

func newMemory() *memory { return &memory{specs: map[string]string{}} }

func (m *memory) Path(slice string) string { return "deltas/specs/" + slice + ".spec.md" }

func (m *memory) Load(_ context.Context, slice string) ([]byte, error) {
	s, ok := m.specs[slice]
	if !ok {
		return nil, errors.New("no such spec")
	}
	return []byte(s), nil
}

func (m *memory) Create(_ context.Context, slice string, c []byte) error {
	if _, ok := m.specs[slice]; ok {
		return errors.New("already exists")
	}
	return m.Save(ctx, slice, c)
}

func (m *memory) Save(_ context.Context, slice string, c []byte) error {
	if m.fail != nil {
		return m.fail
	}
	m.specs[slice] = string(c)
	return nil
}

type tmpl string

func (t tmpl) Spec() ([]byte, error) { return []byte(t), nil }

const tpl = "---\nformat: 1\nslice: {{.Slice}}\nstatus: draft\n---\n\n# {{.Slice}}\n"

const valid = "---\nformat: 1\nslice: billing\nstatus: draft\n---\n- INV-01 [none]\n  When x, y.\n"

func TestCreateMakesADraftFromTheTemplate(t *testing.T) {
	m := newMemory()
	path, err := New(m, tmpl(tpl)).Create(ctx, "billing")
	if err != nil || path != "deltas/specs/billing.spec.md" {
		t.Fatalf("Create = %q, %v", path, err)
	}
	sp, ds := specfile.ParseSlice(path, "billing", []byte(m.specs["billing"]))
	if len(ds) != 0 || sp.Status != "draft" || sp.Slice != "billing" {
		t.Fatalf("created spec is %+v with %v", sp, ds)
	}
}

func TestCreateDoesNotOverwrite(t *testing.T) {
	m := newMemory()
	a := New(m, tmpl(tpl))
	if _, err := a.Create(ctx, "billing"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create(ctx, "billing"); err == nil {
		t.Fatal("a second Create overwrote the spec")
	}
}

func TestCreateReportsABrokenTemplate(t *testing.T) {
	if _, err := New(newMemory(), tmpl("{{.Missing")).Create(ctx, "a"); err == nil {
		t.Fatal("a broken template was accepted")
	}
}

func TestWriteSavesAValidSpec(t *testing.T) {
	m := newMemory()
	ds, err := New(m, tmpl(tpl)).Write(ctx, "billing", []byte(valid))
	if err != nil || len(ds) != 0 || m.specs["billing"] != valid {
		t.Fatalf("Write = %v, %v; stored %q", ds, err, m.specs["billing"])
	}
}

func TestWriteRefusesAnInvalidSpecAndWritesNothing(t *testing.T) {
	cases := map[string]string{
		"bad claim":      strings.Replace(valid, "[none]", "[bogus]", 1),
		"wrong slice":    strings.Replace(valid, "slice: billing", "slice: payments", 1),
		"no frontmatter": "just text",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			m := newMemory()
			ds, err := New(m, tmpl(tpl)).Write(ctx, "billing", []byte(content))
			if err != nil || len(ds) == 0 {
				t.Fatalf("Write = %v, %v; want diagnostics", ds, err)
			}
			if len(m.specs) != 0 {
				t.Fatalf("an invalid spec was stored: %v", m.specs)
			}
			if ds[0].File != "deltas/specs/billing.spec.md" {
				t.Errorf("diagnostics are labeled %q", ds[0].File)
			}
		})
	}
}

func TestWriteSurfacesStoreFailures(t *testing.T) {
	m := newMemory()
	m.fail = errors.New("disk full")
	if _, err := New(m, tmpl(tpl)).Write(ctx, "billing", []byte(valid)); err == nil {
		t.Fatal("a store failure was swallowed")
	}
}

func TestLinkChangesOnlyTheTag(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = valid
	if err := New(m, tmpl(tpl)).Link(ctx, "billing", "INV-01", "  go test ./billing  "); err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(valid, "[none]", "[test: go test ./billing]", 1)
	if m.specs["billing"] != want {
		t.Fatalf("got %q, want %q", m.specs["billing"], want)
	}
}

func TestLinkRejectsBadInput(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = valid
	a := New(m, tmpl(tpl))
	cases := map[string][3]string{
		"bad id":        {"billing", "X-1", "go test ./x"},
		"unknown claim": {"billing", "INV-09", "go test ./x"},
		"empty command": {"billing", "INV-01", "   "},
		"unknown slice": {"ghost", "INV-01", "go test ./x"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if err := a.Link(ctx, c[0], c[1], c[2]); err == nil {
				t.Fatal("accepted")
			}
			if m.specs["billing"] != valid {
				t.Fatalf("the spec changed to %q", m.specs["billing"])
			}
		})
	}
}

func withStatus(status string) string {
	return strings.Replace(valid, "status: draft", "status: "+status, 1)
}

func TestWriteAcceptsOnlyADraftForANewSpec(t *testing.T) {
	m := newMemory()
	ds, err := New(m, tmpl(tpl)).Write(ctx, "billing", []byte(withStatus("approved")))
	if err != nil || len(ds) != 1 || ds[0].Code != CodeNewStatus || len(m.specs) != 0 {
		t.Fatalf("Write = %v, %v; stored %v", ds, err, m.specs)
	}
}

func TestWriteCannotMoveAnExistingSpecForward(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = valid
	ds, err := New(m, tmpl(tpl)).Write(ctx, "billing", []byte(withStatus("verified")))
	if err != nil || len(ds) != 1 || ds[0].Code != CodeSetStatus || m.specs["billing"] != valid {
		t.Fatalf("Write = %v, %v; stored %q", ds, err, m.specs["billing"])
	}
}

func TestWriteSendsAChangedApprovedContractBackToDraft(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = withStatus("approved")
	changed := strings.Replace(withStatus("approved"), "When x, y.", "When x, z.", 1)
	ds, err := New(m, tmpl(tpl)).Write(ctx, "billing", []byte(changed))
	if err != nil || len(ds) != 1 || ds[0].Code != CodeReapproval || !ds[0].Warn {
		t.Fatalf("Write = %v, %v; want one reapproval warning", ds, err)
	}
	if want := strings.Replace(changed, "status: approved", "status: draft", 1); m.specs["billing"] != want {
		t.Fatalf("stored %q, want %q", m.specs["billing"], want)
	}
}

func TestWriteKeepsTheStatusWhenTheContractIsTheSame(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = withStatus("built")
	ds, err := New(m, tmpl(tpl)).Write(ctx, "billing", []byte(withStatus("built")))
	if err != nil || len(ds) != 0 || m.specs["billing"] != withStatus("built") {
		t.Fatalf("Write = %v, %v; stored %q", ds, err, m.specs["billing"])
	}
}

func TestApproveMovesADraftAndTouchesNothingElse(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = valid
	ds, err := New(m, tmpl(tpl)).Approve(ctx, "billing")
	if err != nil || len(ds) != 0 || m.specs["billing"] != withStatus("approved") {
		t.Fatalf("Approve = %v, %v; stored %q", ds, err, m.specs["billing"])
	}
}

func TestApproveRefusesWhatIsNotAnApprovableDraft(t *testing.T) {
	cases := map[string]string{
		"already approved": withStatus("approved"),
		"already built":    withStatus("built"),
		"no claims":        "---\nformat: 1\nslice: billing\nstatus: draft\n---\n",
	}
	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			m := newMemory()
			m.specs["billing"] = spec
			if ds, err := New(m, tmpl(tpl)).Approve(ctx, "billing"); err == nil || len(ds) != 0 || m.specs["billing"] != spec {
				t.Fatalf("Approve = %v, %v; stored %q", ds, err, m.specs["billing"])
			}
		})
	}
}

func TestApproveReportsASpecWithProblems(t *testing.T) {
	m := newMemory()
	broken := strings.Replace(valid, "[none]", "[bogus]", 1)
	m.specs["billing"] = broken
	ds, err := New(m, tmpl(tpl)).Approve(ctx, "billing")
	if err != nil || len(ds) == 0 || m.specs["billing"] != broken {
		t.Fatalf("Approve = %v, %v; stored %q", ds, err, m.specs["billing"])
	}
}

func TestApproveNeedsTheSpec(t *testing.T) {
	if _, err := New(newMemory(), tmpl(tpl)).Approve(ctx, "ghost"); err == nil {
		t.Fatal("a missing spec was approved")
	}
}

func TestBuiltFollowsApprovedOrStaleOnly(t *testing.T) {
	for status, ok := range map[string]bool{"draft": false, "approved": true, "built": false, "verified": false, "stale": true} {
		m := newMemory()
		m.specs["billing"] = withStatus(status)
		_, err := New(m, tmpl(tpl)).Built(ctx, "billing")
		if (err == nil) != ok {
			t.Errorf("Built from %s: err = %v, want ok = %v", status, err, ok)
		}
		if ok && m.specs["billing"] != withStatus("built") {
			t.Errorf("from %s stored %q", status, m.specs["billing"])
		}
	}
}
