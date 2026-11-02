package executor

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
	cmd.WaitDelay = 5 * time.Second
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
