// Package harness embeds what delta init installs into a project. It sits beside the files
// because go:embed cannot reach a parent directory.
package harness

import "embed"

// FS holds core/, skills/ and rules/. Names starting with "." or "_" are left out by embed,
// which is why the git hooks live in core/githooks and are installed as .githooks.
//
//go:embed core skills rules
var FS embed.FS
