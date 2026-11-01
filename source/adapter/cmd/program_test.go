package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/budget"
	"github.com/ronz204/deltas-ai/source/domain/linkcheck"
	"github.com/ronz204/deltas-ai/source/usecase/author"
	"github.com/ronz204/deltas-ai/source/usecase/guard"
	"github.com/ronz204/deltas-ai/source/usecase/measure"
	"github.com/ronz204/deltas-ai/source/usecase/setup"
	"github.com/ronz204/deltas-ai/source/usecase/status"
	"github.com/ronz204/deltas-ai/source/usecase/verify"
)

// root is the project directory the guard reasons about; nothing is read from it.
var root = filepath.Join(os.TempDir(), "delta-cmd-test")

const specTemplate = "---\nformat: 1\nslice: {{.Slice}}\nstatus: draft\n---\n"

const head = "---\nformat: 1\nslice: billing\nstatus: approved\n---\n"

func claim(id, tag string) string { return "- " + id + " [" + tag + "]\n  When x, y.\n" }

// memory stands in for every port the use cases need.
type memory struct {
	specs    map[string]string
	files    []linkcheck.File
	exits    map[string]int
	project  map[string][]byte // files read and written by path, like the git hook
	hooksDir string
	kb       []budget.File // what the always-loaded files are
	onPath   string        // where the delta on the PATH is; empty means none
	self     bool
}

func newMemory() *memory {
	return &memory{specs: map[string]string{}, exits: map[string]int{}, project: map[string][]byte{}}
}

func (m *memory) GitHook(string) ([]byte, error) { return []byte("#!/bin/sh\n"), nil }

func (m *memory) Locate(string) (string, bool) { return m.onPath, m.self }

func (m *memory) Glob(context.Context, []string) ([]budget.File, error) { return m.kb, nil }

func (m *memory) ReadFile(_ context.Context, rel string) ([]byte, error) {
	if b, ok := m.project[rel]; ok {
		return b, nil
	}
	return nil, fs.ErrNotExist
}

func (m *memory) WriteFile(_ context.Context, rel string, c []byte, _ fs.FileMode) error {
	m.project[rel] = c
	return nil
}

func (m *memory) HooksPath(context.Context) (string, error)      { return m.hooksDir, nil }
func (m *memory) SetHooksPath(_ context.Context, p string) error { m.hooksDir = p; return nil }

func (m *memory) Path(s string) string { return "deltas/" + s + ".spec.md" }

func (m *memory) List(context.Context) ([]string, error) {
	return slices.Sorted(maps.Keys(m.specs)), nil
}

func (m *memory) Load(_ context.Context, s string) ([]byte, error) {
	src, ok := m.specs[s]
	if !ok {
		return nil, fmt.Errorf("no spec for %s", s)
	}
	return []byte(src), nil
}

func (m *memory) Create(ctx context.Context, s string, c []byte) error {
	if _, ok := m.specs[s]; ok {
		return errors.New("already exists")
	}
	return m.Save(ctx, s, c)
}

func (m *memory) Save(_ context.Context, s string, c []byte) error {
	m.specs[s] = string(c)
	return nil
}

func (m *memory) Spec() ([]byte, error) { return []byte(specTemplate), nil }

func (m *memory) TestFiles(context.Context) ([]linkcheck.File, error) { return m.files, nil }

func (m *memory) Run(_ context.Context, command string) (int, error) { return m.exits[command], nil }

func (m *memory) run(stdin string, args ...string) (out, errOut string, err error) {
	installer := setup.New(m, m, m, m)
	cmd := New(Deps{
		Author:    author.New(m, m),
		Checker:   verify.New(m, m, m, nil),
		Report:    status.New(m),
		Guard:     guard.New(root, status.New(m), installer, nil),
		Installer: installer,
		Meter:     measure.New(m),
		Version:   "test",
	})
	var o, e bytes.Buffer
	cmd.SetArgs(args)
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetOut(&o)
	cmd.SetErr(&e)
	err = cmd.ExecuteContext(context.Background())
	return o.String(), e.String(), err
}

func TestSpecNewCreatesADraftAndSaysWhere(t *testing.T) {
	m := newMemory()
	out, _, err := m.run("", "spec", "new", "billing")
	if err != nil || out != "created deltas/billing.spec.md\n" {
		t.Fatalf("out = %q, err = %v", out, err)
	}
	if !strings.Contains(m.specs["billing"], "slice: billing") {
		t.Fatalf("stored %q", m.specs["billing"])
	}
	if _, _, err := m.run("", "spec", "new", "billing"); err == nil || errors.Is(err, ErrFound) {
		t.Fatalf("a second new = %v, want a plain error", err)
	}
}

func TestSpecWriteSendsFindingsToStderr(t *testing.T) {
	m := newMemory()
	out, errOut, err := m.run("just text", "spec", "write", "billing")
	if !errors.Is(err, ErrFound) || out != "" {
		t.Fatalf("err = %v, stdout = %q", err, out)
	}
	if !strings.Contains(errOut, "SPEC-FRONTMATTER") || !strings.Contains(errOut, "nothing was written") || len(m.specs) != 0 {
		t.Fatalf("stderr = %q, stored %v", errOut, m.specs)
	}

	out, errOut, err = m.run(head+claim("INV-01", "none"), "spec", "write", "billing")
	if err != nil || out != "wrote the spec of billing\n" || errOut != "" {
		t.Fatalf("out = %q, errOut = %q, err = %v", out, errOut, err)
	}
}

func TestSpecLintFailsOnAnErrorAndStaysQuietOtherwise(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = head + claim("INV-01", "none")
	m.specs["broken"] = "no frontmatter"
	_, errOut, err := m.run("", "spec", "lint")
	if !errors.Is(err, ErrFound) || !strings.Contains(errOut, "deltas/broken.spec.md:1: SPEC-FRONTMATTER") {
		t.Fatalf("err = %v, stderr = %q", err, errOut)
	}
	out, errOut, err := m.run("", "spec", "lint", "billing")
	if err != nil || out != "" || errOut != "" {
		t.Fatalf("a clean spec printed %q %q, err = %v", out, errOut, err)
	}
}

func TestLintWarningsDoNotFailButErrorsDo(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = head + claim("INV-01", "test: go test ./a")
	m.files = []linkcheck.File{{Path: "a/a_test.go", Text: "// INV-01\nt.Skip(\"later\")\n"}}
	_, errOut, err := m.run("", "lint")
	if err != nil || !strings.Contains(errOut, "warning LINT-SKIP") {
		t.Fatalf("err = %v, stderr = %q; a warning alone must not fail", err, errOut)
	}
	m.files = nil
	_, errOut, err = m.run("", "lint")
	if !errors.Is(err, ErrFound) || !strings.Contains(errOut, "LINT-NO-CITE") {
		t.Fatalf("err = %v, stderr = %q", err, errOut)
	}
}

func TestCheckPrintsOneRowPerClaimAndASummary(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = head + claim("INV-01", "test: go test ./a") + claim("INV-02", "test: go test ./b") + claim("INV-03", "none")
	m.files = []linkcheck.File{{Path: "a/a_test.go", Text: "// INV-01"}, {Path: "b/b_test.go", Text: "// INV-02"}}
	m.exits["go test ./b"] = 1

	out, _, err := m.run("", "check")
	if !errors.Is(err, ErrFound) {
		t.Fatalf("a violated claim must fail: %v", err)
	}
	var rows []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		rows = append(rows, strings.Join(strings.Fields(line), " "))
	}
	want := []string{
		"billing INV-01 Holds",
		"billing INV-02 Violated the command exited 1",
		"billing INV-03 Unverified nothing verifies this claim yet",
		"1 holds, 1 violated, 1 unverified",
	}
	if strings.Join(rows, "|") != strings.Join(want, "|") {
		t.Fatalf("rows:\n%s\nwant:\n%s", strings.Join(rows, "\n"), strings.Join(want, "\n"))
	}
	for _, line := range strings.Split(out, "\n") {
		if line != strings.TrimRight(line, " ") {
			t.Fatalf("trailing spaces in %q", line)
		}
	}

	m.exits["go test ./b"] = 0
	if _, _, err := m.run("", "check"); err != nil {
		t.Fatalf("nothing violated must pass: %v", err)
	}
}

func TestStatusTable(t *testing.T) {
	m := newMemory()
	out, _, err := m.run("", "status")
	if err != nil || !strings.Contains(out, "delta spec new") {
		t.Fatalf("an empty project = %q, %v", out, err)
	}
	m.specs["billing"] = head + claim("INV-01", "test: x") + claim("INV-02", "manual") + claim("INV-03", "none")
	out, _, err = m.run("", "status")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if err != nil || len(lines) != 2 ||
		strings.Join(strings.Fields(lines[0]), " ") != "SLICE STATE CLAIMS TEST MANUAL NONE ISSUES" ||
		strings.Join(strings.Fields(lines[1]), " ") != "billing approved 3 1 1 1 0" {
		t.Fatalf("status = %q, %v", out, err)
	}
}

func TestLinkConfirmsAndChangesTheTag(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = head + claim("INV-01", "none")
	out, _, err := m.run("", "link", "billing", "INV-01", "go test ./billing")
	if err != nil || out != "linked INV-01 of billing\n" {
		t.Fatalf("out = %q, err = %v", out, err)
	}
	if !strings.Contains(m.specs["billing"], "- INV-01 [test: go test ./billing]") {
		t.Fatalf("stored %q", m.specs["billing"])
	}
}

func TestCommitMsgReadsTheFileGitGives(t *testing.T) {
	dir := t.TempDir()
	good, bad := filepath.Join(dir, "good"), filepath.Join(dir, "bad")
	if err := os.WriteFile(good, []byte("feat(cli): add check, lint and status commands\n# comment\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte("Fixed it.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := newMemory()
	if out, errOut, err := m.run("", "commit-msg", good); err != nil || out != "" || errOut != "" {
		t.Fatalf("a good message printed %q %q, err = %v", out, errOut, err)
	}
	_, errOut, err := m.run("", "commit-msg", bad)
	if !errors.Is(err, ErrFound) || !strings.Contains(errOut, bad+":1: COMMIT-") {
		t.Fatalf("err = %v, stderr = %q", err, errOut)
	}
	if _, _, err := m.run("", "commit-msg", filepath.Join(dir, "missing")); err == nil || errors.Is(err, ErrFound) {
		t.Fatalf("a missing file = %v, want a plain error", err)
	}
}

func TestWrongArgumentCountsAreUsageErrors(t *testing.T) {
	for _, args := range [][]string{{"spec", "new"}, {"spec", "write"}, {"link", "billing", "INV-01"}, {"lint", "x"}, {"check", "a", "b"}, {"status", "x"}, {"commit-msg"}, {"nope"}} {
		if _, _, err := newMemory().run("", args...); err == nil || errors.Is(err, ErrFound) {
			t.Errorf("%v = %v, want a usage error", args, err)
		}
	}
}

func TestHelpListsTheCommandsAndNoCompletion(t *testing.T) {
	out, _, err := newMemory().run("", "--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"spec", "link", "lint", "check", "status", "commit-msg"} {
		if !strings.Contains(out, name) {
			t.Errorf("help does not list %q", name)
		}
	}
	if strings.Contains(out, "completion") {
		t.Error("the completion command should be disabled")
	}
}
