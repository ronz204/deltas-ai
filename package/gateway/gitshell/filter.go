package gitshell

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

const sniffed = 8000

func (g *Git) counts(name string) bool { return !g.ignore(name) && !g.binary(name) }

func (g *Git) binary(name string) bool {
	f, err := os.Open(filepath.Join(g.dir, filepath.FromSlash(name)))
	if err != nil {
		return false
	}
	defer f.Close()
	head := make([]byte, sniffed)
	n, _ := io.ReadFull(f, head)
	return bytes.IndexByte(head[:n], 0) >= 0
}
