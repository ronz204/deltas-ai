// Package gitcli talks to the git binary.
package gitcli

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Git runs git inside one directory.
type Git struct{ dir string }

func New(dir string) *Git { return &Git{dir: dir} }

// HooksPath returns the core.hooksPath git would use here, from any config level, or "" when it is not set.
func (g *Git) HooksPath(ctx context.Context) (string, error) {
	out, err := g.run(ctx, "config", "--get", "core.hooksPath")
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 { // git config --get exits 1 when the key is unset
		return "", nil
	}
	return strings.TrimSpace(out), err
}

// SetHooksPath sets core.hooksPath for this repository only.
func (g *Git) SetHooksPath(ctx context.Context, path string) error {
	_, err := g.run(ctx, "config", "--local", "core.hooksPath", path)
	return err
}

func (g *Git) run(ctx context.Context, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", g.dir}, args...)...).Output()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return string(out), fmt.Errorf("git %s: %s: %w", args[0], strings.TrimSpace(string(exit.Stderr)), err)
	}
	return string(out), err
}
