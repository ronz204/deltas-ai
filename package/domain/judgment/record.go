package judgment

type Mark struct {
	Tree     string
	Violated bool
}

type Standing int

const (
	Unchecked Standing = iota
	Failing
	Passing
)

func (m Mark) At(tree string) Standing {
	switch {
	case m.Tree == "" || m.Tree != tree:
		return Unchecked
	case m.Violated:
		return Failing
	}
	return Passing
}
