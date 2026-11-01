package conffile

import (
	"reflect"
	"testing"
	"time"

	"github.com/ronz204/deltas-ai/source/domain/budget"
)

func TestAnEmptyFileChangesNothing(t *testing.T) {
	c := Parse("")
	if c.Budget != budget.Defaults() || len(c.TestFiles) != 0 || c.Timeout != 0 || len(c.Tools) != 0 || len(c.Scopes) != 0 || len(c.Problems) != 0 {
		t.Fatalf("Parse(\"\") = %+v", c)
	}
}

func TestParseReadsEveryKey(t *testing.T) {
	c := Parse("# settings\r\n\r\n" +
		"always_tokens = 100\n" +
		"test_files=*_spec.rb, *Test.java\n" +
		"command_timeout=90s\n" +
		"required_tools=go git\n" +
		"scope.billing=./billing/... shared/money/\n" +
		"scope.payments = payments\n")
	if c.Budget.Total != 100 || c.Budget.Router != 80 {
		t.Errorf("budget = %+v", c.Budget)
	}
	if want := []string{"*_spec.rb", "*Test.java"}; !reflect.DeepEqual(c.TestFiles, want) {
		t.Errorf("test_files = %v, want %v", c.TestFiles, want)
	}
	if c.Timeout != 90*time.Second {
		t.Errorf("timeout = %v", c.Timeout)
	}
	if want := []string{"go", "git"}; !reflect.DeepEqual(c.Tools, want) {
		t.Errorf("tools = %v, want %v", c.Tools, want)
	}
	if want := map[string][]string{"billing": {"billing", "shared/money"}, "payments": {"payments"}}; !reflect.DeepEqual(c.Scopes, want) {
		t.Errorf("scopes = %v, want %v", c.Scopes, want)
	}
	if len(c.Problems) != 0 {
		t.Errorf("problems = %v", c.Problems)
	}
}

func TestALineThatCannotBeUsedIsToldAndLeftOut(t *testing.T) {
	c := Parse("nonsense\n" +
		"colour=blue\n" +
		"test_files=[\n" +
		"command_timeout=soon\n" +
		"command_timeout=-5m\n" +
		"scope.billing=.\n" +
		"scope.=billing\n" +
		"required_tools=go\n")
	if len(c.TestFiles) != 0 || c.Timeout != 0 || len(c.Scopes) != 0 {
		t.Errorf("a bad line took effect: %+v", c)
	}
	if len(c.Tools) != 1 {
		t.Errorf("the good line after the bad ones was lost: %v", c.Tools)
	}
	want := []string{
		`deltas/deltas.config:1: expected key=value`,
		`deltas/deltas.config:2: unknown key "colour"`,
		`deltas/deltas.config:3: test_files: "[" is not a valid pattern`,
		`deltas/deltas.config:4: command_timeout: "soon" is not a positive duration such as 10m`,
		`deltas/deltas.config:5: command_timeout: "-5m" is not a positive duration such as 10m`,
		`deltas/deltas.config:6: scope.<slice> needs a slice name and at least one path inside the project`,
		`deltas/deltas.config:7: scope.<slice> needs a slice name and at least one path inside the project`,
	}
	if !reflect.DeepEqual(c.Problems, want) {
		t.Errorf("problems:\n got %q\nwant %q", c.Problems, want)
	}
}

func TestTheLastValueOfAKeyWins(t *testing.T) {
	c := Parse("required_tools=a\nrequired_tools=b\nscope.x=one\nscope.x=two\n")
	if !reflect.DeepEqual(c.Tools, []string{"b"}) || !reflect.DeepEqual(c.Scopes["x"], []string{"two"}) {
		t.Fatalf("Parse = %+v", c)
	}
}
