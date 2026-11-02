package conffile

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/budget"
)

func TestNothingToReadMeansTheDefaults(t *testing.T) {
	c := Parse("", "")
	if c.Budget != budget.Defaults() || len(c.Files.Match) != 0 || len(c.Files.Skip) != 0 || len(c.Problems) != 0 {
		t.Fatalf("an empty config = %+v, want the defaults", c)
	}
}

func TestTheBudgetKeysOverrideTheDefaults(t *testing.T) {
	c := Parse("# a comment\nalways_tokens=900\nclaude_lines=40\ndoc_lines=300\n", "")
	if c.Budget.Total != 900 || c.Budget.Router != 40 || c.Budget.Rule != budget.Defaults().Rule || len(c.Problems) != 0 {
		t.Fatalf("budget = %+v, problems = %v", c.Budget, c.Problems)
	}
}

func TestTestingConfSaysWhatATestIsAndWhereNotToLook(t *testing.T) {
	c := Parse("", "# the patterns\nmatch=*_spec.rb, *Test.java\nskip=dist,.venv\n")
	if want := []string{"*_spec.rb", "*Test.java"}; !reflect.DeepEqual(c.Files.Match, want) {
		t.Errorf("match = %v, want %v", c.Files.Match, want)
	}
	if want := []string{"dist", ".venv"}; !reflect.DeepEqual(c.Files.Skip, want) {
		t.Errorf("skip = %v, want %v", c.Files.Skip, want)
	}
	if len(c.Problems) != 0 {
		t.Errorf("problems = %v", c.Problems)
	}
}

func TestALineThatCannotBeUsedIsLeftOutAndTold(t *testing.T) {
	c := Parse("colour=blue\nalways_tokens\n", "match=[\nsize=3\nskip\n")
	want := []string{
		`deltas/budget.conf:1: unknown key "colour"`,
		`deltas/budget.conf:2: expected key=value`,
		`deltas/testing.conf:1: match: "[" is not a valid pattern`,
		`deltas/testing.conf:2: unknown key "size"`,
		`deltas/testing.conf:3: expected key=value`,
	}
	if !reflect.DeepEqual(c.Problems, want) {
		t.Fatalf("problems:\n%s\nwant:\n%s", strings.Join(c.Problems, "\n"), strings.Join(want, "\n"))
	}
	if len(c.Files.Match) != 0 {
		t.Errorf("a bad pattern was kept: %v", c.Files.Match)
	}
}
