package judgment

import "testing"

func TestMarkAt(t *testing.T) {
	cases := []struct {
		name string
		mark Mark
		tree string
		want Standing
	}{
		{"never checked", Mark{}, "a", Unchecked},
		{"an empty tree never matches an empty mark", Mark{}, "", Unchecked},
		{"another tree", Mark{Tree: "a"}, "b", Unchecked},
		{"another tree, even if it failed", Mark{Tree: "a", Violated: true}, "b", Unchecked},
		{"same tree, violated", Mark{Tree: "a", Violated: true}, "a", Failing},
		{"same tree, nothing violated", Mark{Tree: "a"}, "a", Passing},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.mark.At(c.tree); got != c.want {
				t.Fatalf("At(%q) = %v, want %v", c.tree, got, c.want)
			}
		})
	}
}
