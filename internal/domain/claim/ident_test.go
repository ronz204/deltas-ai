package claim

import "testing"

func TestParseID(t *testing.T) {
	good := []struct {
		in   string
		kind Kind
	}{
		{"INV-01", Invariant},
		{"INV-1", Invariant},
		{"INV-120", Invariant},
		{"ACC-01", Acceptance},
	}
	for _, c := range good {
		id, err := ParseID(c.in)
		if err != nil || id.String() != c.in || id.Kind() != c.kind {
			t.Errorf("ParseID(%q) = %v (kind %v), %v", c.in, id, id.Kind(), err)
		}
	}
	for _, bad := range []string{"", "INV", "INV-", "INV-x", "inv-01", "FOO-01", "INV-01 ", " INV-01", "INV_01", "INV-01a", "DSN-01"} {
		if _, err := ParseID(bad); err == nil {
			t.Errorf("ParseID(%q) accepted a bad ID", bad)
		}
	}
	if (ID{}).Kind() != 0 {
		t.Error("the zero ID must have no kind")
	}
}
