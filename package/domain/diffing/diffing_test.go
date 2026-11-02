package diffing

import (
	"reflect"
	"testing"
)

func TestChanged(t *testing.T) {
	before := map[string]string{"a.go": "1", "b.go": "1", "c.go": "1"}
	cases := map[string]struct {
		after map[string]string
		want  []string
	}{
		"nothing changed": {map[string]string{"a.go": "1", "b.go": "1", "c.go": "1"}, nil},
		"edited":          {map[string]string{"a.go": "2", "b.go": "1", "c.go": "1"}, []string{"a.go"}},
		"added":           {map[string]string{"a.go": "1", "b.go": "1", "c.go": "1", "d.go": "1"}, []string{"d.go"}},
		"removed":         {map[string]string{"a.go": "1", "c.go": "1"}, []string{"b.go"}},
		"all three":       {map[string]string{"a.go": "2", "c.go": "1", "d.go": "1"}, []string{"a.go", "b.go", "d.go"}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := Changed(before, c.after); !reflect.DeepEqual(got, c.want) {
				t.Fatalf("Changed = %v, want %v", got, c.want)
			}
		})
	}
}

func TestChangedWithNothingBeforeIsEverything(t *testing.T) {
	if got := Changed(nil, map[string]string{"a.go": "1"}); !reflect.DeepEqual(got, []string{"a.go"}) {
		t.Fatalf("Changed = %v", got)
	}
}
