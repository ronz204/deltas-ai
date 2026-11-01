package judgment

// Mark is what delta check leaves for a slice: the tree it looked at and whether a claim was violated.
type Mark struct {
	Tree     string
	Violated bool
}

// Standing is what a mark proves about the tree as it is now.
type Standing int

const (
	Unchecked Standing = iota // no check has seen this tree
	Failing                   // a check saw this tree and a claim was violated
	Passing                   // a check saw this tree and no claim was violated
)

// At judges the mark against the current tree; an empty mark, or one from another tree, proves nothing.
func (m Mark) At(tree string) Standing {
	switch {
	case m.Tree == "" || m.Tree != tree:
		return Unchecked
	case m.Violated:
		return Failing
	}
	return Passing
}
