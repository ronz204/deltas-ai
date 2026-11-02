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

func (g *Git) run(ctx context.Context, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", g.dir}, args...)...).Output()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return string(out), fmt.Errorf("git %s: %s: %w", args[0], strings.TrimSpace(string(exit.Stderr)), err)
	}
	return string(out), err
}
