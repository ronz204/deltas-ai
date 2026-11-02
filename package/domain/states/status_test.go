package states

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

func TestCanAllowsOnlyTheMovesOfTheCycle(t *testing.T) {
	legal := map[[2]Status]bool{
		{Draft, Approved}: true,
		{Approved, Built}: true,
		{Built, Verified}: true,
		{Verified, Stale}: true,
		{Stale, Built}:    true,
		{Approved, Draft}: true,
		{Built, Draft}:    true,
		{Verified, Draft}: true,
		{Stale, Draft}:    true,
	}
	for _, from := range statuses {
		for _, to := range statuses {
			if got := Can(from, to); got != legal[[2]Status{from, to}] {
				t.Errorf("Can(%s, %s) = %v, want %v", from, to, got, !got)
			}
		}
	}
}

func TestCanRefusesAnUnknownStatus(t *testing.T) {
	if Can("", Draft) || Can("done", Draft) || Can("", Approved) {
		t.Error("an unknown status moved")
	}
}
