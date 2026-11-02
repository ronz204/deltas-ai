package gitshell

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/linking"
)

type Git struct {
	dir    string
	ignore func(string) bool
}

func New(dir string) *Git { return &Git{dir: dir, ignore: linking.IgnoreMatcher(nil)} }

func (g *Git) Ignoring(patterns []string) *Git {
	return &Git{dir: g.dir, ignore: linking.IgnoreMatcher(patterns)}
}

func (g *Git) run(ctx context.Context, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", g.dir}, args...)...).Output()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return string(out), fmt.Errorf("git %s: %s: %w", args[0], strings.TrimSpace(string(exit.Stderr)), err)
	}
	return string(out), err
}
