// Package manifest reads and writes deltas.manifest, the record of the harness files delta put in a project.
package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
)

// Name is where the manifest lives, relative to the project root.
const Name = "deltas/deltas.manifest"

const header = "# deltas.manifest: written by delta, not by hand. The SHA-256 of each file delta installed, with LF line endings.\n"

// Manifest says, for each file delta installed and keeps up to date, the hash it had when delta wrote it.
type Manifest struct {
	Files map[string]string // project path with forward slashes -> Hash of its content
}

// Hash is the SHA-256 of content with CRLF read as LF, so a checkout with other line endings is the same file.
func Hash(content []byte) string {
	sum := sha256.Sum256([]byte(strings.ReplaceAll(string(content), "\r\n", "\n")))
	return hex.EncodeToString(sum[:])
}

// Parse reads "<hash>  <path>" lines. Blank lines, lines starting with # and any line that is not of that shape
// are skipped: a file whose record is lost is only one that delta no longer knows it installed.
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

// Format writes the manifest with its paths sorted, so the same files give the same text.
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
