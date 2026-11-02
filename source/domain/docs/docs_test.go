package docs

import "testing"

func TestFilledOnlyWhenNoGuideCommentIsLeft(t *testing.T) {
	cases := map[string]bool{
		"# Overview\n\nA real doc.\n":                  true,
		"# Overview\n\n<!-- what the project is -->\n": false,
		"": true,
	}
	for text, want := range cases {
		if got := Filled(text); got != want {
			t.Errorf("Filled(%q) = %v, want %v", text, got, want)
		}
	}
}

func TestNameOnlyForADocFileDirectlyUnderTheFolder(t *testing.T) {
	cases := map[string]struct {
		name string
		ok   bool
	}{
		"deltas/docs/overview.md": {"overview", true},
		"deltas/docs/deep/a.md":   {"", false},
		"deltas/docs/notes.txt":   {"", false},
		"deltas/specs/a.spec.md":  {"", false},
		"docs/overview.md":        {"", false},
	}
	for p, want := range cases {
		if name, ok := Name(p); name != want.name && ok || ok != want.ok {
			t.Errorf("Name(%q) = %q, %v; want %q, %v", p, name, ok, want.name, want.ok)
		}
	}
}
