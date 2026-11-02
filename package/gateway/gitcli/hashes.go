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

func (g *Git) Tree(ctx context.Context, scope linking.Scope) (hash string, dirty bool, err error) {
	head, changed, err := g.head(ctx)
	if err != nil {
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
	fmt.Fprintf(sum, "%s\n", head)
	for _, name := range names {
		fmt.Fprintf(sum, "%s\x00%s\n", name, g.contentHash(name))
	}
	return hex.EncodeToString(sum.Sum(nil)), true, nil
}

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

func (g *Git) head(ctx context.Context) (head string, changed []string, err error) {
	out, err := g.run(ctx, "rev-parse", "--verify", "-q", "HEAD")
	var exit *exec.ExitError
	switch {
	case err == nil:
		return strings.TrimSpace(out), []string{"diff", "--name-only", "--no-renames", "-z", "HEAD"}, nil
	case errors.As(err, &exit) && exit.ExitCode() == 1:
		return "", []string{"ls-files", "-z", "--cached"}, nil
	}
	return "", nil, err
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
