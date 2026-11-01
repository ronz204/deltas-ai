package main

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The same claims with the same evidence must get the same verdicts in every language.
var wantVerdicts = map[string]string{
	"INV-01": "Holds",
	"INV-02": "Violated",
	"INV-03": "Unverified", // the command exits 0 but no test cites it
	"INV-04": "Holds",      // a skipped test exits 0: the known limit, which lint warns about
	"INV-05": "Unverified",
	"ACC-01": "Unverified",
}

func copyFixture(t *testing.T, name string) *project {
	t.Helper()
	p := newProject(t)
	src := filepath.Join("testdata", "compat", name)
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(p.root, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func goIsAvailable() string {
	if _, err := exec.LookPath("go"); err != nil {
		return "the go tool is not on the PATH"
	}
	return ""
}

func nodeRunsTypeScript() string {
	out, err := exec.Command("node", "-p", "process.features.typescript").Output()
	if err != nil {
		return "node is not installed"
	}
	if feature := strings.TrimSpace(string(out)); feature == "false" || feature == "undefined" {
		return "this node cannot run TypeScript files"
	}
	return ""
}

func TestCompat(t *testing.T) {
	if testing.Short() {
		t.Skip("the compatibility suite runs real test runners")
	}
	cases := []struct {
		name, fixture, uncited string
		unavailable            func() string
	}{
		{"go", "go", "uncited/uncited_test.go", goIsAvailable},
		{"typescript", "ts", "uncited/uncited.test.ts", nodeRunsTypeScript},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if why := c.unavailable(); why != "" {
				t.Skip(why)
			}
			p := copyFixture(t, c.fixture)

			code, out, _ := p.delta("", "check")
			if code != 1 {
				t.Fatalf("check exited %d, want 1 because INV-02 is violated\n%s", code, out)
			}
			for id, want := range wantVerdicts {
				if got := verdict(out, "compat", id); got != want {
					t.Errorf("%s = %q, want %q\n%s", id, got, want, out)
				}
			}
			if !strings.Contains(out, "2 holds, 1 violated, 3 unverified") {
				t.Errorf("summary missing:\n%s", out)
			}

			code, _, errOut := p.delta("", "lint")
			if code != 1 || !strings.Contains(errOut, "LINT-NO-CITE") || !strings.Contains(errOut, "warning LINT-SKIP") {
				t.Errorf("lint = %d, want 1 with LINT-NO-CITE and a LINT-SKIP warning\n%s", code, errOut)
			}

			p.write(c.uncited, p.read(c.uncited)+"\n// INV-03\n")
			_, out, _ = p.delta("", "check")
			if got := verdict(out, "compat", "INV-03"); got != "Holds" {
				t.Errorf("after the test cites INV-03 it is %q, want Holds\n%s", got, out)
			}
		})
	}
}
