// Package bundle hands out the files embedded in the harness.
package bundle

import "github.com/ronz204/deltas-ai/source/harness"

type Bundle struct{}

func (Bundle) Spec() ([]byte, error) { return harness.FS.ReadFile("core/templates/spec.md") }
