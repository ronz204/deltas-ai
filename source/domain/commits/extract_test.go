package commits

import "testing"

func TestFromCommandReadsTheMessage(t *testing.T) {
	cases := []struct {
		name, command, want string
	}{
		{"double quotes", `git commit -m "feat(x): add the thing"`, "feat(x): add the thing"},
		{"single quotes", `git commit -m 'feat(x): add the thing'`, "feat(x): add the thing"},
		{"unquoted word", `git commit -m fix`, "fix"},
		{"long flag", `git commit --message "feat(x): add it"`, "feat(x): add it"},
		{"long flag with equals", `git commit --message="feat(x): add it"`, "feat(x): add it"},
		{"glued to the flag", `git commit -m"feat(x): add it"`, "feat(x): add it"},
		{"combined short flags", `git commit -am "feat(x): add it"`, "feat(x): add it"},
		{"flags around it", `git commit -q --no-verify -m "feat(x): add it" --signoff`, "feat(x): add it"},
		{"two paragraphs", `git commit -m "feat(x): add it" -m "The body."`, "feat(x): add it\n\nThe body."},
		{"newlines inside the quotes", "git commit -m \"feat(x): add it\n\nThe body.\"", "feat(x): add it\n\nThe body."},
		{"after other commands", `git add a.go && git diff --cached | tail -1 && git commit -m "feat(x): add it"`, "feat(x): add it"},
		{"after an assignment line", "C=\"x\"\ngit add .\ngit commit -m \"feat(x): add it\"", "feat(x): add it"},
		{"with git options", `git -C sub -c user.name=a commit -m "feat(x): add it"`, "feat(x): add it"},
		{"escaped quote inside", `git commit -m "feat(x): say \"hi\""`, `feat(x): say "hi"`},
		{"dollar escaped", `git commit -m "cost \$5"`, "cost $5"},
		{"a variable stays literal", `git commit -m "feat(x): add it $C"`, "feat(x): add it $C"},
		{"heredoc", "git commit -m \"$(cat <<'EOF'\nfeat(x): add it\n\nThe body.\nEOF\n)\"", "feat(x): add it\n\nThe body."},
		{"heredoc with quotes in the body", "git commit -m \"$(cat <<'EOF'\nfeat(x): say \"hi\"\n\nIt's (fine).\nEOF\n)\"", "feat(x): say \"hi\"\n\nIt's (fine)."},
		{"heredoc tag inside a longer word", "git commit -m \"$(cat <<'EOF'\nfeat(x): add it\nEOFS stays in the body\nEOF\n)\"", "feat(x): add it\nEOFS stays in the body"},
		{"heredoc with unquoted tag", "git commit -m \"$(cat <<EOF\nfeat(x): add it\nEOF\n)\"", "feat(x): add it"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := FromCommand(c.command)
			if !ok || got != c.want {
				t.Fatalf("FromCommand(%q) = %q, %v; want %q", c.command, got, ok, c.want)
			}
		})
	}
}

func TestFromCommandGivesUpWhenThereIsNoMessageToRead(t *testing.T) {
	for _, command := range []string{
		"",
		"ls -la",
		"git status",
		"git commit",
		"git commit --amend",
		"git commit -m",
		"git commit -F msg.txt",
		"git commit --file=msg.txt",
		"git commit -F - <<'EOF'\nfeat(x): add it\nEOF",
		`git commit -m "$(date)"`,
		"git tag -a v1 -m \"$(cat <<'EOF'\nv1 release\nEOF\n)\"",
		`echo "git commit -m feat"`,
		`git log -m`,
		`git push && echo commit -m x`,
	} {
		if got, ok := FromCommand(command); ok {
			t.Errorf("FromCommand(%q) = %q, want no message", command, got)
		}
	}
}
