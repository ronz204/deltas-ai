package commit

import (
	"reflect"
	"strings"
	"testing"
)

func codes(message string) []string {
	var out []string
	for _, d := range Validate("MSG", message) {
		out = append(out, d.Code)
	}
	return out
}

func TestValidateAcceptsRealSubjects(t *testing.T) {
	good := []string{
		"feat(domain): add spec parser and claim rules",
		"docs(elements): mark first Phase 1 tasks done",
		"chore(repo): remove unused ROADMAP.md template",
		"build(go): add module and delta entry point",
		"test(arch): enforce layer dependency rules",
		"fix(skills): fix step 7 break and surveyor text",
		"feat(domain): add spec parser and claim rules\n\nThe body explains why.\n\nCo-Authored-By: Someone <a@b.c>",
		"feat(domain): add spec parser and claim rules\r\n\r\nbody",
		"# a leading comment\nfeat(domain): add spec parser and claim rules\n# Please enter the commit message\n",
		"\n\nfeat(domain): add spec parser and claim rules\n",
		"feat: add a parser that reads the whole spec",
	}
	for _, m := range good {
		if got := codes(m); len(got) != 0 {
			t.Errorf("rejected %q: %v", m, got)
		}
	}
}

func TestValidateRejects(t *testing.T) {
	cases := []struct {
		name, message string
		want          []string
	}{
		{"empty", "", []string{CodeEmpty}},
		{"only comments", "# nothing\n# here\n", []string{CodeEmpty}},
		{"no type", "add the spec parser and the claim rules ok", []string{CodeFormat}},
		{"no space after colon", "feat(domain):add spec parser and claim rules", []string{CodeFormat}},
		{"unknown type", "wip(domain): add spec parser and claim rules", []string{CodeType}},
		{"capital type", "Feat(domain): add spec parser and claim rules", []string{CodeFormat}},
		{"too short", "feat(domain): add parser", []string{CodeLength}},
		{"too long", "feat(domain): add the spec parser and the claim rules now", []string{CodeLength}},
		{"capital subject", "feat(domain): Add spec parser and claim rules", []string{CodeCase}},
		{"trailing period", "feat(domain): add spec parser and claim rules.", []string{CodePeriod}},
		{"body glued to the subject", "feat(domain): add spec parser and claim rules\nbody", []string{CodeBody}},
		{"several problems", "Wip: Add parser.", []string{CodeLength, CodeFormat}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := codes(c.message); !reflect.DeepEqual(got, c.want) {
				t.Fatalf("codes = %v, want %v", got, c.want)
			}
		})
	}
}

func TestValidateLetsGitGeneratedMessagesThrough(t *testing.T) {
	for _, m := range []string{"Merge branch 'main' into feature", `Revert "feat(domain): add spec parser"`, "fixup! feat(domain): add spec parser", "squash! x"} {
		if got := codes(m); len(got) != 0 {
			t.Errorf("%q should not be validated: %v", m, got)
		}
	}
}

func TestValidateCountsCharactersNotBytes(t *testing.T) {
	// 40 characters but more than 50 bytes
	m := "fix(áéí): " + strings.Repeat("ñ", 30)
	if got := codes(m); len(got) != 0 {
		t.Fatalf("a 40 character subject was rejected: %v", got)
	}
}

func TestValidateReportsTheSubjectLine(t *testing.T) {
	ds := Validate("MSG", "# c\n\nfeat(domain): Add spec parser and claim rules\n")
	if len(ds) != 1 || ds[0].Line != 3 || ds[0].File != "MSG" {
		t.Fatalf("got %+v, want one diagnostic on line 3", ds)
	}
}
