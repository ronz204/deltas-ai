package executor

import (
	"os"
	"os/exec"
)

func (r *Runner) Locate(name string) (path string, self bool) {
	found, err := exec.LookPath(name)
	if err != nil {
		return "", false
	}
	own, err := os.Executable()
	if err != nil {
		return found, false
	}
	a, errA := os.Stat(found)
	b, errB := os.Stat(own)
	return found, errA == nil && errB == nil && os.SameFile(a, b)
}
