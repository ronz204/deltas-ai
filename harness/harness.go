// Package harness embeds what delta init installs; it sits beside the files because go:embed cannot go up.
package harness

import "embed"

// FS holds core/, skills/ and rules/; embed skips names starting with "." or "_", so the git hooks live in core/githooks.
//
//go:embed core skills rules
var FS embed.FS
