package specfile

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/package/domain/claimset"
	"github.com/ronz204/deltas-ai/package/domain/findings"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

const head = "---\nformat: 1\nslice: invoicing\nstatus: approved\n---\n"

const validSpec = head + `
# Invoicing: spec

## Invariants
- INV-01 [test: go test ./invoicing]
  When an invoice is issued, the system shall round each line to 2 decimals
  before summing.
- INV-02 [none]
  The system never issues two invoices with the same sequence number.
- INV-03 [manual: read the PDF of a sample invoice]
  Where a tax rate is set, tax is computed on the discounted subtotal.

## Acceptance
- ACC-01 [test: go test ./e2e/sequence]
  Given invoice 1001 exists, when another is issued with 1001, then it is rejected.
`

func codes(ds []findings.Diagnostic) []string {
	out := make([]string, len(ds))
	for i, d := range ds {
		out[i] = d.Code
	}
	return out
}

func TestParseValid(t *testing.T) {
	sp, ds := Parse("invoicing.spec.md", []byte(validSpec))
	if len(ds) != 0 {
		t.Fatalf("unexpected diagnostics: %v", ds)
	}
	if sp.Format != 1 || sp.Slice != "invoicing" || sp.Status != states.Approved {
		t.Fatalf("frontmatter read as %+v", sp)
	}
	id := func(s string) claimset.ID {
		v, err := claimset.ParseID(s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	want := []claimset.Claim{
		{ID: id("INV-01"), Tag: claimset.Tag{Method: claimset.Test, Arg: "go test ./invoicing"}, Text: "When an invoice is issued, the system shall round each line to 2 decimals before summing.", Line: 10},
		{ID: id("INV-02"), Tag: claimset.Tag{Method: claimset.None}, Text: "The system never issues two invoices with the same sequence number.", Line: 13},
		{ID: id("INV-03"), Tag: claimset.Tag{Method: claimset.Manual, Arg: "read the PDF of a sample invoice"}, Text: "Where a tax rate is set, tax is computed on the discounted subtotal.", Line: 15},
		{ID: id("ACC-01"), Tag: claimset.Tag{Method: claimset.Test, Arg: "go test ./e2e/sequence"}, Text: "Given invoice 1001 exists, when another is issued with 1001, then it is rejected.", Line: 19},
	}
	if !reflect.DeepEqual(sp.Claims, want) {
		t.Fatalf("claims:\n got %+v\nwant %+v", sp.Claims, want)
	}
}

func TestParseSpanish(t *testing.T) {
	src := "---\nformat: 1\nslice: facturacion\nstatus: draft\n---\n" +
		"- INV-01 [test: go test ./facturacion]\n  Cuando se emite una factura, el sistema debe redondear cada línea a 2 decimales.\n" +
		"- ACC-01 [none]\n  Dado que existe la factura 1001, cuando se intenta emitir otra con 1001, entonces se rechaza.\n"
	sp, ds := Parse("facturacion.spec.md", []byte(src))
	if len(ds) != 0 || len(sp.Claims) != 2 {
		t.Fatalf("got %d claims, diagnostics %v", len(sp.Claims), ds)
	}
}

func TestParseSameWhateverTheLineEnding(t *testing.T) {
	want, wantDs := Parse("a", []byte(validSpec))
	for name, src := range map[string]string{
		"crlf": strings.ReplaceAll(validSpec, "\n", "\r\n"),
		"bom":  "\xef\xbb\xbf" + validSpec,
	} {
		got, gotDs := Parse("a", []byte(src))
		if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(gotDs, wantDs) {
			t.Errorf("%s: parsed differently from the LF version", name)
		}
	}
}

func TestParseKeepsBracketsInsideACommand(t *testing.T) {
	src := head + "- INV-01 [test: go test -run 'a[b]' ./...]\n  The system is fast.\n"
	sp, ds := Parse("a", []byte(src))
	if len(ds) != 0 || len(sp.Claims) != 1 || sp.Claims[0].Tag.Arg != "go test -run 'a[b]' ./..." {
		t.Fatalf("got %+v, %v", sp.Claims, ds)
	}
}

func TestParseIgnoresWhatIsNotAClaim(t *testing.T) {
	src := head + "```markdown\n- INV-01 [none]\n  Shown as an example only.\n```\n" +
		"<!-- - INV-02 [none]\n  Also an example. -->\n" +
		"<!-- one line: - INV-03 [none] -->\n" +
		"- a plain bullet\n- another one\n  indented\n" +
		"- INV-04 [none]\n  The only real claim.\n"
	sp, ds := Parse("a", []byte(src))
	if len(ds) != 0 || len(sp.Claims) != 1 || sp.Claims[0].ID.String() != "INV-04" {
		t.Fatalf("got %+v, %v", sp.Claims, ds)
	}
}

func TestParseDiagnostics(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		codes []string
		line  int // line of the first diagnostic
		kept  int // claims that still parse cleanly beside the problem
	}{
		{"empty file", "", []string{CodeFrontmatter}, 1, 0},
		{"no frontmatter", "# Title\n", []string{CodeFrontmatter}, 1, 0},
		{"unclosed frontmatter", "---\nformat: 1\n", []string{CodeFrontmatter}, 1, 0},
		{"unknown key", "---\nformat: 1\nslice: a\nstatus: draft\nowner: me\n---\n", []string{CodeKey}, 5, 0},
		{"duplicate key", "---\nformat: 1\nformat: 1\nslice: a\nstatus: draft\n---\n", []string{CodeKey}, 3, 0},
		{"malformed key line", "---\nformat 1\nslice: a\nstatus: draft\n---\n", []string{CodeKey, CodeKey}, 2, 0},
		{"missing keys", "---\n---\n", []string{CodeKey, CodeKey, CodeKey}, 1, 0},
		{"future format", "---\nformat: 2\nslice: a\nstatus: draft\n---\n", []string{CodeFormat}, 2, 0},
		{"format not a number", "---\nformat: one\nslice: a\nstatus: draft\n---\n", []string{CodeFormat}, 2, 0},
		{"bad slice name", "---\nformat: 1\nslice: Invoicing\nstatus: draft\n---\n", []string{CodeSlice}, 3, 0},
		{"bad status", "---\nformat: 1\nslice: a\nstatus: done\n---\n", []string{CodeStatus}, 4, 0},
		{"header without tag", head + "- INV-01\n  When x, y.\n", []string{CodeClaim}, 6, 0},
		{"header with unknown prefix", head + "- FOO-01 [none]\n  When x, y.\n", []string{CodeClaim}, 6, 0},
		{"lowercase id", head + "- inv-01 [none]\n  When x, y.\n", []string{CodeClaim}, 6, 0},
		{"retired check tag", head + "- INV-01 [check: ./x.sh]\n  When x, y.\n", []string{CodeTag}, 6, 0},
		{"test tag without command", head + "- INV-01 [test]\n  When x, y.\n", []string{CodeTag}, 6, 0},
		{"duplicate id", head + "- INV-01 [none]\n  When x, y.\n- INV-01 [none]\n  When x, z.\n", []string{CodeDupID}, 8, 1},
		{"no sentence", head + "- INV-01 [none]\n\nprose\n", []string{CodeNoText}, 6, 0},
		{"sentence not indented", head + "- INV-01 [none]\nWhen x, y.\n", []string{CodeNoText}, 6, 0},
		{"invariant not EARS", head + "- INV-01 [none]\n  If x, y.\n", []string{CodeStatement}, 6, 0},
		{"acceptance not Given/When/Then", head + "- ACC-01 [none]\n  The suite passes.\n", []string{CodeStatement}, 6, 0},
		{"bad tag and bad sentence", head + "- INV-01 [bogus]\n  When x\n", []string{CodeTag, CodeStatement}, 6, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sp, ds := Parse("a.spec.md", []byte(c.src))
			if !reflect.DeepEqual(codes(ds), c.codes) {
				t.Fatalf("codes = %v, want %v (%v)", codes(ds), c.codes, ds)
			}
			if ds[0].Line != c.line || ds[0].File != "a.spec.md" {
				t.Errorf("first diagnostic at %s:%d, want line %d", ds[0].File, ds[0].Line, c.line)
			}
			if len(sp.Claims) != c.kept {
				t.Errorf("kept %d claims, want %d", len(sp.Claims), c.kept)
			}
		})
	}
}

func TestParseKeepsGoodClaimsBesideBadOnes(t *testing.T) {
	src := head + "- INV-01 [none]\n  When x, y.\n- INV-02 [bogus]\n  When x, y.\n- INV-03 [none]\n  When x, z.\n"
	sp, ds := Parse("a", []byte(src))
	if len(ds) != 1 || len(sp.Claims) != 2 || sp.Claims[0].ID.String() != "INV-01" || sp.Claims[1].ID.String() != "INV-03" {
		t.Fatalf("got %+v, %v", sp.Claims, ds)
	}
}
