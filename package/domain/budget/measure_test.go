package budget

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func lines(n int) string { return strings.Repeat("x\n", n) }

func TestParseCeilings(t *testing.T) {
	if got := ParseCeilings(""); got != Defaults() {
		t.Fatalf("an empty conf = %+v, want the defaults %+v", got, Defaults())
	}
	conf := "# a comment\n\nalways_tokens=100\nclaude_lines = 10\r\nrule_always_lines=0\nskill_desc_chars=abc\nagent_desc_chars=-5\nreference_lines=150\nnonsense\n"
	want := Ceilings{Total: 100, Router: 10, Rule: 0, Skill: 1024, Agent: 1024}
	if got := ParseCeilings(conf); got != want {
		t.Fatalf("got %+v, want %+v (0 turns a ceiling off; bad values and other keys change nothing)", got, want)
	}
}

func TestReadMatter(t *testing.T) {
	cases := []struct {
		name string
		text string
		want matter
	}{
		{"none", "# Title\nbody\n", matter{}},
		{"one-line description", "---\nname: a\ndescription: Does a thing.\n---\nbody\n", matter{end: 4, desc: len("Does a thing.")}},
		{"folded description", "---\nname: a\ndescription: >-\n  first line\n  second line\n---\nbody\n", matter{end: 6, desc: len("first line") + 1 + len("second line") + 1}},
		{"literal description", "---\ndescription: |\n  abc\n---\n", matter{end: 4, desc: len("abc") + 1}},
		{"with paths", "---\npaths:\n  - \"src/**\"\n---\nbody\n", matter{end: 4, paths: true}},
		{"paths after a description", "---\ndescription: x\npaths:\n  - a\n---\n", matter{end: 5, paths: true, desc: 1}},
		{"crlf", "---\r\ndescription: abc\r\n---\r\nbody\r\n", matter{end: 3, desc: 3}},
		{"counts characters, not bytes", "---\ndescription: ñandú\n---\n", matter{end: 3, desc: 5}},
		{"an unclosed block never ends", "---\ndescription: x\nbody\n", matter{desc: 1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ls := strings.Split(strings.ReplaceAll(strings.TrimSuffix(c.text, "\n"), "\r", ""), "\n")
			if got := readMatter(ls); got != c.want {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestMeasureKeepsOnlyWhatLoadsOnEveryTurn(t *testing.T) {
	files := []File{
		{"CLAUDE.md", "# Project\n" + lines(9)},
		{".claude/rules/always.md", "# Rule\n" + lines(4)},
		{".claude/rules/scoped.md", "---\npaths:\n  - \"src/**\"\n---\n" + lines(40)},
		{".claude/skills/drafter/SKILL.md", "---\nname: drafter\ndescription: Drafts the contract.\n---\n" + lines(200)},
		{".claude/skills/drafter/references/guide.md", lines(300)},
		{".claude/agents/reviewer.md", "---\nname: reviewer\ndescription: Reviews code.\n---\n" + lines(30)},
		{".claude/docs/overview.md", lines(100)},
		{"README.md", lines(100)},
	}
	rep := Measure(files, Defaults())
	var got []string
	for _, it := range rep.Items {
		got = append(got, it.Kind+":"+it.Path)
	}
	want := []string{
		"router:CLAUDE.md",
		"rule:.claude/rules/always.md",
		"skill:.claude/skills/drafter/SKILL.md",
		"agent:.claude/agents/reviewer.md",
	}
	slices.Sort(got)
	slices.Sort(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("items = %v, want exactly %v", got, want)
	}
}

func TestMeasureSizesAndTokens(t *testing.T) {
	skillText := "---\nname: drafter\ndescription: Drafts the contract.\n---\n" + lines(200)
	files := []File{
		{"CLAUDE.md", "# Project\n" + lines(9)},
		{".claude/skills/drafter/SKILL.md", skillText},
	}
	rep := Measure(files, Defaults())
	byKind := map[string]Item{}
	for _, it := range rep.Items {
		byKind[it.Kind] = it
	}
	if r := byKind["router"]; r.Size != 10 || r.Unit != "lines" || r.Limit != 80 || r.Tokens != len("# Project\n"+lines(9))*10/33 {
		t.Errorf("router = %+v", r)
	}
	if s := byKind["skill"]; s.Size != len("Drafts the contract.") || s.Unit != "chars" || s.Limit != 1024 || s.Tokens != len("Drafts the contract.")*10/33 {
		t.Errorf("skill = %+v: only the description is loaded, not the body", s)
	}
	if rep.Total != byKind["router"].Tokens+byKind["skill"].Tokens || rep.Limit != 3500 {
		t.Errorf("total = %d (limit %d)", rep.Total, rep.Limit)
	}
}

func TestMeasureOrdersTheHeaviestFirst(t *testing.T) {
	files := []File{
		{".claude/rules/b.md", lines(3)},
		{".claude/rules/a.md", lines(3)},
		{".claude/rules/c.md", lines(30)},
	}
	var paths []string
	for _, it := range Measure(files, Defaults()).Items {
		paths = append(paths, it.Path)
	}
	if want := []string{".claude/rules/c.md", ".claude/rules/a.md", ".claude/rules/b.md"}; !reflect.DeepEqual(paths, want) {
		t.Fatalf("order = %v, want %v (by tokens, then by path)", paths, want)
	}
}

func TestMeasureOfNothingOrAnEmptyFile(t *testing.T) {
	if rep := Measure(nil, Defaults()); len(rep.Items) != 0 || rep.Total != 0 {
		t.Fatalf("nothing = %+v", rep)
	}
	rep := Measure([]File{{"CLAUDE.md", ""}}, Defaults())
	if len(rep.Items) != 1 || rep.Items[0].Size != 0 {
		t.Fatalf("an empty router = %+v", rep.Items)
	}
}

func TestFindingsNameEveryCeilingThatIsPassed(t *testing.T) {
	files := []File{
		{"CLAUDE.md", lines(81)},
		{".claude/rules/a.md", lines(51)},
		{".claude/rules/ok.md", lines(50)},
	}
	rep := Measure(files, Defaults())
	var got []string
	for _, d := range rep.Findings() {
		got = append(got, d.Code+" "+d.File)
	}
	if want := []string{"BUDGET-ITEM CLAUDE.md", "BUDGET-ITEM .claude/rules/a.md"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("findings = %v, want %v (50 lines is at the ceiling, not over it)", got, want)
	}

	rep = Measure(files, Ceilings{Total: 10})
	ds := rep.Findings()
	if len(ds) != 1 || ds[0].Code != CodeTotal {
		t.Fatalf("only the total ceiling is set: %v", ds)
	}
	if ds = Measure(files, Ceilings{}).Findings(); len(ds) != 0 {
		t.Fatalf("with no ceilings nothing can be passed: %v", ds)
	}
}
