package claimset

type Claim struct {
	ID   ID
	Tag  Tag
	Text string // the statement, continuation lines joined by single spaces
	Line int    // 1-based line of its header in the spec
}
