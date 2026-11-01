package slice

import "testing"

func TestParseStatus(t *testing.T) {
	for _, st := range statuses {
		got, err := ParseStatus(string(st))
		if err != nil || got != st {
			t.Errorf("ParseStatus(%q) = %q, %v", st, got, err)
		}
	}
	for _, bad := range []string{"", "Draft", "done", " draft"} {
		if _, err := ParseStatus(bad); err == nil {
			t.Errorf("ParseStatus(%q) accepted a bad status", bad)
		}
	}
}
