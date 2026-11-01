package spec

import (
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/internal/domain/claim"
)

func mustID(t *testing.T, s string) claim.ID {
	t.Helper()
	id, err := claim.ParseID(s)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestSetTagChangesOnlyThatLine(t *testing.T) {
	tag := claim.Tag{Kind: claim.Test, Arg: "go test ./invoicing"}
	got, err := SetTag([]byte(validSpec), mustID(t, "INV-02"), tag)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(validSpec, "- INV-02 [none]", "- INV-02 [test: go test ./invoicing]", 1)
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	sp, ds := Parse("a", got)
	if len(ds) != 0 || sp.Claims[1].Tag != tag {
		t.Fatalf("the edited spec does not parse back to the new tag: %+v, %v", sp.Claims, ds)
	}
}

func TestSetTagKeepsCRLFAndBOM(t *testing.T) {
	src := "\xef\xbb\xbf" + strings.ReplaceAll(validSpec, "\n", "\r\n")
	got, err := SetTag([]byte(src), mustID(t, "INV-02"), claim.Tag{Kind: claim.Manual, Arg: "read it"})
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(src, "- INV-02 [none]", "- INV-02 [manual: read it]", 1)
	if string(got) != want {
		t.Fatalf("line endings or BOM changed:\n%q\nwant\n%q", got, want)
	}
}

func TestSetTagUnknownClaim(t *testing.T) {
	if _, err := SetTag([]byte(validSpec), mustID(t, "INV-99"), claim.Tag{Kind: claim.None}); err == nil {
		t.Fatal("an unknown claim was accepted")
	}
	broken := head + "- INV-01 [bogus]\n  When x, y.\n"
	if _, err := SetTag([]byte(broken), mustID(t, "INV-01"), claim.Tag{Kind: claim.None}); err == nil {
		t.Fatal("a claim with problems must not be editable")
	}
}

func TestParseSliceChecksTheName(t *testing.T) {
	if _, ds := ParseSlice("f", "invoicing", []byte(validSpec)); len(ds) != 0 {
		t.Fatalf("matching name rejected: %v", ds)
	}
	_, ds := ParseSlice("f", "billing", []byte(validSpec))
	if len(ds) != 1 || ds[0].Code != CodeName {
		t.Fatalf("got %v, want one %s", ds, CodeName)
	}
}

func TestValidSlice(t *testing.T) {
	for _, ok := range []string{"a", "invoicing", "order-flow", "v2", "a-1-b"} {
		if !ValidSlice(ok) {
			t.Errorf("%q rejected", ok)
		}
	}
	for _, bad := range []string{"", "Invoicing", "a_b", "-a", "a-", "a--b", "../x", "a/b", `a\b`, "a b", "a."} {
		if ValidSlice(bad) {
			t.Errorf("%q accepted", bad)
		}
	}
}
