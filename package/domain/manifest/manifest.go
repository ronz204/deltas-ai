package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
)

const Name = "deltas/deltas.manifest"

const header = "# deltas.manifest: written by delta, not by hand. The SHA-256 of each file delta installed, with LF line endings.\n"

type Manifest struct {
	Files map[string]string
}

func Hash(content []byte) string {
	sum := sha256.Sum256([]byte(strings.ReplaceAll(string(content), "\r\n", "\n")))
	return hex.EncodeToString(sum[:])
}

func Parse(text string) Manifest {
	m := Manifest{Files: map[string]string{}}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		hash, path, ok := strings.Cut(line, "  ")
		if !ok || len(hash) != sha256.Size*2 || strings.TrimSpace(path) == "" {
			continue
		}
		m.Files[strings.TrimSpace(path)] = hash
	}
	return m
}

func (m Manifest) Format() string {
	paths := make([]string, 0, len(m.Files))
	for p := range m.Files {
		paths = append(paths, p)
	}
	slices.Sort(paths)
	var b strings.Builder
	b.WriteString(header)
	for _, p := range paths {
		fmt.Fprintf(&b, "%s  %s\n", m.Files[p], p)
	}
	return b.String()
}
