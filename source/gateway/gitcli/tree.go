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

	"github.com/ronz204/deltas-ai/package/domain/linking"
)

// Tree fingerprints the part of the working tree a scope covers: HEAD plus the content of every file in it
// that differs from HEAD or is untracked and not ignored. dirty is false when there is nothing of the sort,
// as right after a commit.
// shortcut: it reads every such file, so a big binary left untracked slows it; upgrade by hashing size and mtime.
func (g *Git) Tree(ctx context.Context, scope linking.Scope) (hash string, dirty bool, err error) {
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
	names = slices.DeleteFunc(names, func(name string) bool { return !scope.Covers(name) })
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

// Files fingerprints, file by file, every file of the project a scope covers: the tracked ones and the
// untracked ones that are not ignored. Unlike Tree it does not compare with HEAD, so it tells what changed
// since an earlier call even after a commit.
// shortcut: it hashes every covered file on each call; upgrade by keeping size and mtime beside the hash.
func (g *Git) Files(ctx context.Context, scope linking.Scope) (map[string]string, error) {
	names, err := g.names(ctx, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	files := make(map[string]string, len(names))
	for _, name := range names {
		if scope.Covers(name) {
			files[name] = g.contentHash(name)
		}
	}
	return files, nil
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
