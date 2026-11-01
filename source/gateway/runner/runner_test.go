package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSplit(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"go test ./...", []string{"go", "test", "./..."}},
		{"  go   test\t./x  ", []string{"go", "test", "./x"}},
		{`go test -run 'a b' ./x`, []string{"go", "test", "-run", "a b", "./x"}},
		{`bunx vitest run "my file.test.ts"`, []string{"bunx", "vitest", "run", "my file.test.ts"}},
		{`run C:\tools\check.exe`, []string{"run", `C:\tools\check.exe`}},
		{`echo ""`, []string{"echo", ""}},
		{`a'b c'd`, []string{"ab cd"}},
		{`say "it's"`, []string{"say", "it's"}},
	}
	for _, c := range cases {
		got, err := split(c.in)
		if err != nil || !reflect.DeepEqual(got, c.want) {
			t.Errorf("split(%q) = %q, %v; want %q", c.in, got, err, c.want)
		}
	}
	for _, bad := range []string{"", "   ", `echo "oops`, `echo 'oops`} {
		if got, err := split(bad); err == nil {
			t.Errorf("split(%q) = %q, want an error", bad, got)
		}
	}
}

// TestHelperProcess is the command the other tests run: the test binary itself, switched by an environment variable.
func TestHelperProcess(t *testing.T) {
	switch os.Getenv("DELTA_HELPER") {
	case "":
		return
	case "exit3":
		fmt.Fprintln(os.Stdout, "to stdout")
		fmt.Fprintln(os.Stderr, "to stderr")
		os.Exit(3)
	case "sleep":
		time.Sleep(time.Minute)
	case "pwd":
		wd, _ := os.Getwd()
		fmt.Println(wd)
	}
}

func helper(t *testing.T, mode string) string {
	t.Helper()
	t.Setenv("DELTA_HELPER", mode)
	return fmt.Sprintf(`"%s" -test.run=^TestHelperProcess$`, os.Args[0])
}

func TestRunReturnsTheExitCodeAndShowsOutput(t *testing.T) {
	var out bytes.Buffer
	r := New(t.TempDir(), time.Minute, &out)
	cases := []struct {
		mode string
		want int
	}{{"", 0}, {"exit3", 3}}
	for _, c := range cases {
		out.Reset()
		got, err := r.Run(context.Background(), helper(t, c.mode), "")
		if err != nil || got != c.want {
			t.Fatalf("mode %q: exit %d, %v; want %d", c.mode, got, err, c.want)
		}
	}
	if !strings.Contains(out.String(), "to stdout") || !strings.Contains(out.String(), "to stderr") {
		t.Fatalf("the command's output was not shown: %q", out.String())
	}
}

func TestRunUsesTheDirectory(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	if _, err := New(dir, time.Minute, &out).Run(context.Background(), helper(t, "pwd"), ""); err != nil {
		t.Fatal(err)
	}
	first, _, _ := strings.Cut(out.String(), "\n")
	got, gotErr := filepath.EvalSymlinks(strings.TrimSpace(first)) // macOS temp dirs sit behind a symlink
	want, wantErr := filepath.EvalSymlinks(dir)
	if gotErr != nil || wantErr != nil || got != want {
		t.Fatalf("ran in %q, want %q (%v, %v)", got, want, gotErr, wantErr)
	}
}

func TestRunErrors(t *testing.T) {
	r := New(t.TempDir(), 300*time.Millisecond, &bytes.Buffer{})
	if _, err := r.Run(context.Background(), "definitely-not-a-command-xyz", ""); err == nil {
		t.Error("a command that cannot start must be an error, not an exit code")
	}
	if _, err := r.Run(context.Background(), "", ""); err == nil {
		t.Error("an empty command must be an error")
	}
	start := time.Now()
	_, err := r.Run(context.Background(), helper(t, "sleep"), "")
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("err = %v, want a timeout", err)
	}
	if time.Since(start) > 30*time.Second {
		t.Fatal("the timeout did not stop the command")
	}
}

func TestRunStartsInTheGivenFolderOfTheProject(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "services", "auth"), 0o755); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if _, err := New(root, time.Minute, &out).Run(context.Background(), helper(t, "pwd"), "services/auth"); err != nil {
		t.Fatal(err)
	}
	first, _, _ := strings.Cut(out.String(), "\n")
	got, gotErr := filepath.EvalSymlinks(strings.TrimSpace(first))
	want, wantErr := filepath.EvalSymlinks(filepath.Join(root, "services", "auth"))
	if gotErr != nil || wantErr != nil || got != want {
		t.Fatalf("ran in %q, want %q (%v, %v)", got, want, gotErr, wantErr)
	}
}
