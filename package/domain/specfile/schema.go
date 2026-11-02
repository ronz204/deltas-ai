package specfile

import (
	"github.com/ronz204/deltas-ai/package/domain/claimset"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

const supportedFormat = 1

type Spec struct {
	Format int
	Slice  string
	Status states.Status
	Claims []claimset.Claim
}

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
