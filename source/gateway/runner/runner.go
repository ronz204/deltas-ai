// Package runner runs the commands that verify claims.
package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

type Runner struct {
	dir     string
	timeout time.Duration
	out     io.Writer
}

func New(dir string, timeout time.Duration, out io.Writer) *Runner {
	return &Runner{dir: dir, timeout: timeout, out: out}
}

// Run returns the exit code; an error means it did not finish (could not start, or timed out). The command
// runs from the project root.
// shortcut: no shell, so no pipes, redirections or $VARIABLES; write such a check as a program or script.
func (r *Runner) Run(ctx context.Context, command string) (int, error) {
	args, err := split(command)
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = r.dir
	cmd.Stdout, cmd.Stderr = r.out, r.out
	cmd.WaitDelay = 5 * time.Second // a killed command's children must not keep the pipes open
	err = cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return 0, fmt.Errorf("timed out after %s", r.timeout)
	}
	var exit *exec.ExitError
	switch {
	case err == nil:
		return 0, nil
	case errors.As(err, &exit):
		return exit.ExitCode(), nil
	}
	return 0, err
}
