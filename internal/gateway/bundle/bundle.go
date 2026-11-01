// Package bundle hands out the files embedded in the harness.
package bundle

import "github.com/ronz204/deltas-ai/harness"

// Bundle is the embedded harness.
type Bundle struct{}

// Spec is the template a new spec starts from.
func (Bundle) Spec() ([]byte, error) { return harness.FS.ReadFile("core/templates/spec.md") }
