package runner

import (
	"os"
	"os/exec"
)

// Locate finds a program on the PATH; path is empty when there is none. self is true when it is
// the program that is running, which tells delta apart from another program with the same name.
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
