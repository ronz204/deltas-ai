package claim

import "testing"

func TestParseTag(t *testing.T) {
	good := []struct {
		in   string
		want Tag
	}{
		{"test: go test ./facturacion", Tag{Test, "go test ./facturacion"}},
		{"test:go test ./x", Tag{Test, "go test ./x"}},
		{"  test :  bunx vitest run  ", Tag{Test, "bunx vitest run"}},
		{`test: go test -run 'a[b]' ./...`, Tag{Test, `go test -run 'a[b]' ./...`}},
		{`test: go run C:\tools\check`, Tag{Test, `go run C:\tools\check`}},
		{"manual", Tag{Manual, ""}},
		{"manual: open the invoice and read the total", Tag{Manual, "open the invoice and read the total"}},
		{"none", Tag{None, ""}},
	}
	for _, c := range good {
		got, err := ParseTag(c.in)
		if err != nil || got != c.want {
			t.Errorf("ParseTag(%q) = %+v, %v; want %+v", c.in, got, err, c.want)
		}
	}
	for _, bad := range []string{"", "test", "test:", "test:   ", "none: why", "check: ./x.sh", "Test: x", "manually", "tests: x"} {
		if _, err := ParseTag(bad); err == nil {
			t.Errorf("ParseTag(%q) accepted a bad tag", bad)
		}
	}
}
