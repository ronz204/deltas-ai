package manifest

import (
	"strings"
	"testing"
)

func TestHashReadsCRLFAsLF(t *testing.T) {
	if Hash([]byte("a\r\nb\r\n")) != Hash([]byte("a\nb\n")) {
		t.Fatal("the same text with other line endings hashed differently")
	}
	if Hash([]byte("a")) == Hash([]byte("b")) {
		t.Fatal("different text hashed alike")
	}
}

func TestFormatThenParseGivesTheSameFilesInOrder(t *testing.T) {
	m := Manifest{Files: map[string]string{".claude/rules/b.md": Hash([]byte("b")), ".claude/rules/a.md": Hash([]byte("a"))}}
	text := m.Format()
	if strings.Index(text, "a.md") > strings.Index(text, "b.md") {
		t.Fatalf("paths are not sorted:\n%s", text)
	}
	got := Parse(text)
	if len(got.Files) != 2 || got.Files[".claude/rules/a.md"] != Hash([]byte("a")) {
		t.Fatalf("Parse(Format) = %v", got.Files)
	}
	if got.Format() != text {
		t.Fatal("formatting twice gave another text")
	}
}

func TestParseSkipsWhatIsNotARecord(t *testing.T) {
	good := Hash([]byte("x")) + "  .claude/rules/x.md"
	m := Parse("# a comment\n\n" + good + "\nnot a record\nabc  short-hash.md\n" + Hash([]byte("y")) + "  \n")
	if len(m.Files) != 1 || m.Files[".claude/rules/x.md"] == "" {
		t.Fatalf("Parse = %v, want only the one good record", m.Files)
	}
}
