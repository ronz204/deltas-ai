package gitcli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// Tree fingerprints the working tree: HEAD plus the content of every file that differs from it or is
// untracked and not ignored. dirty is false when there is nothing of the sort, as right after a commit.
// shortcut: it reads every such file, so a big binary left untracked slows it; upgrade by hashing size and mtime.
func (g *Git) Tree(ctx context.Context) (hash string, dirty bool, err error) {
	head, err := g.run(ctx, "rev-parse", "--verify", "-q", "HEAD")
	var exit *exec.ExitError
	changed := []string{"diff", "--name-only", "--no-renames", "-z", "HEAD"}
	switch {
	case errors.As(err, &exit) && exit.ExitCode() == 1: // a repository with no commit yet
		head, changed = "", []string{"ls-files", "-z", "--cached"}
	case err != nil:
		return "", false, err
	}
	names, err := g.names(ctx, changed...)
	if err != nil {
		return "", false, err
	}
	untracked, err := g.names(ctx, "ls-files", "-z", "--others", "--exclude-standard")
	if err != nil {
		return "", false, err
	}
	names = append(names, untracked...)
	slices.Sort(names)
	names = slices.Compact(names)
	if len(names) == 0 {
		return "", false, nil
	}

	sum := sha256.New()
	fmt.Fprintf(sum, "%s\n", strings.TrimSpace(head))
	for _, name := range names {
		fmt.Fprintf(sum, "%s\x00%s\n", name, g.contentHash(name))
	}
	return hex.EncodeToString(sum.Sum(nil)), true, nil
}

func (g *Git) names(ctx context.Context, args ...string) ([]string, error) {
	out, err := g.run(ctx, args...)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, n := range strings.Split(out, "\x00") {
		if n != "" {
			names = append(names, n)
		}
	}
	return names, nil
}

func (g *Git) contentHash(name string) string {
	text, err := os.ReadFile(filepath.Join(g.dir, filepath.FromSlash(name)))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "deleted"
	case err != nil:
		return "unreadable"
	}
	sum := sha256.Sum256(text)
	return hex.EncodeToString(sum[:])
}
