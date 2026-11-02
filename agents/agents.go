// Package agents embeds what delta init installs; it sits beside the files because go:embed cannot go up.
package agents

import "embed"

// FS holds core/, docs/, skills/ and rules/; embed skips names starting with "." or "_", so the git hooks live in core/githooks.
//
//go:embed core docs skills rules
var FS embed.FS
