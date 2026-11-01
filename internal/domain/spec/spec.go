// Package spec reads a slice's contract file: a frontmatter block, then claims.
package spec

import (
	"github.com/ronz204/deltas-ai/internal/domain/claim"
	"github.com/ronz204/deltas-ai/internal/domain/slice"
)

// supportedFormat is the only spec format this version reads. Bumping it is how a future
// version tells old specs apart and migrates them.
const supportedFormat = 1

// Spec is a parsed contract.
type Spec struct {
	Format int
	Slice  string
	Status slice.Status
	Claims []claim.Claim
}

// Diagnostic codes. Hooks and CI depend on these, so they do not change once released.
const (
	CodeFrontmatter = "SPEC-FRONTMATTER" // missing or unclosed --- block
	CodeKey         = "SPEC-KEY"         // malformed, unknown, duplicate or missing frontmatter key
	CodeFormat      = "SPEC-FORMAT"      // format is not one this version reads
	CodeSlice       = "SPEC-SLICE"       // slice name is not lowercase words joined by hyphens
	CodeStatus      = "SPEC-STATUS"      // status is not a known one
	CodeClaim       = "SPEC-CLAIM"       // claim header is malformed or its ID is not valid
	CodeTag         = "SPEC-TAG"         // verification tag is not valid
	CodeDupID       = "SPEC-DUP-ID"      // claim ID already defined
	CodeNoText      = "SPEC-NO-TEXT"     // claim has no sentence under its header
	CodeStatement   = "SPEC-STATEMENT"   // sentence does not follow EARS or Given/When/Then
	CodeName        = "SPEC-NAME"        // frontmatter names a different slice than the file
)
