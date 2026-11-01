// Package specfile reads a slice's spec file: a frontmatter block, then claims.
package specfile

import (
	"github.com/ronz204/deltas-ai/internal/domain/assertion"
	"github.com/ronz204/deltas-ai/internal/domain/lifecycle"
)

// supportedFormat is the only format this version reads; bumping it is how old specs get migrated.
const supportedFormat = 1

type Spec struct {
	Format int
	Slice  string
	Status lifecycle.Status
	Claims []assertion.Claim
}

// Diagnostic codes. Hooks and CI depend on these, so they do not change once released.
const (
	CodeFrontmatter = "SPEC-FRONTMATTER"
	CodeKey         = "SPEC-KEY"
	CodeFormat      = "SPEC-FORMAT"
	CodeSlice       = "SPEC-SLICE"
	CodeStatus      = "SPEC-STATUS"
	CodeClaim       = "SPEC-CLAIM"
	CodeTag         = "SPEC-TAG"
	CodeDupID       = "SPEC-DUP-ID"
	CodeNoText      = "SPEC-NO-TEXT"
	CodeStatement   = "SPEC-STATEMENT"
	CodeName        = "SPEC-NAME"
)
