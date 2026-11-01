package commits

import (
	"regexp"
	"strings"
)

var (
	heredocOpen = regexp.MustCompile(`\bgit\b[^\n;&|]*\bcommit\b[^\n;&|]*?(?:-\w*m|--message)\s*=?\s*"\$\(cat <<-?\s*['"]?(\w+)['"]?\n`)
	shortFlags  = regexp.MustCompile(`^-[a-zA-Z]*m$`)
)

// FromCommand returns the message a command line gives to git commit with -m or --message. It
// returns false when there is none it can read: no commit, an editor, -F, or a message built by
// a shell substitution other than the usual heredoc.
func FromCommand(command string) (string, bool) {
	if message, ok := fromHeredoc(command); ok {
		return message, true
	}
	for _, words := range segments(command) {
		if message, ok := messageOf(words); ok {
			return message, true
		}
	}
	return "", false
}

// fromHeredoc reads the message of git commit -m "$(cat <<'TAG' ... TAG)" straight from the text,
// because the quotes inside the body would confuse the word splitter.
func fromHeredoc(command string) (string, bool) {
	loc := heredocOpen.FindStringSubmatchIndex(command)
	if loc == nil {
		return "", false
	}
	tag := command[loc[2]:loc[3]]
	body := command[loc[1]:]
	for from := 0; ; {
		at := strings.Index(body[from:], "\n"+tag)
		if at < 0 {
			return "", false
		}
		end := from + at + 1 + len(tag)
		if end == len(body) || strings.ContainsRune(")\r\n \t", rune(body[end])) {
			return body[:from+at], true
		}
		from += at + 1
	}
}

func messageOf(words []string) (string, bool) {
	if len(words) == 0 || words[0] != "git" {
		return "", false
	}
	i := 1
	for ; i < len(words) && strings.HasPrefix(words[i], "-"); i++ {
		if words[i] == "-C" || words[i] == "-c" {
			i++
		}
	}
	if i >= len(words) || words[i] != "commit" {
		return "", false
	}

	var paragraphs []string
	args := words[i+1:]
	for j := 0; j < len(args); j++ {
		switch a := args[j]; {
		case a == "-F" || a == "--file" || strings.HasPrefix(a, "--file="):
			return "", false
		case a == "--message" || shortFlags.MatchString(a):
			if j+1 >= len(args) {
				return "", false
			}
			j++
			paragraphs = append(paragraphs, args[j])
		case strings.HasPrefix(a, "--message="):
			paragraphs = append(paragraphs, strings.TrimPrefix(a, "--message="))
		case strings.HasPrefix(a, "-m") && !strings.HasPrefix(a, "--"):
			paragraphs = append(paragraphs, a[2:])
		}
	}
	message := strings.Join(paragraphs, "\n\n")
	if len(paragraphs) == 0 || strings.HasPrefix(message, "$(") {
		return "", false
	}
	return message, true
}

// segments splits a shell command line into the words of each simple command, honouring quotes.
// A newline, ;, | and & end a command unless they are quoted.
func segments(command string) [][]string {
	var out [][]string
	var words []string
	var word strings.Builder
	var quote rune
	started := false // true once a word began, so "" still counts as one

	endWord := func() {
		if started {
			words = append(words, word.String())
			word.Reset()
			started = false
		}
	}
	endCommand := func() {
		endWord()
		if len(words) > 0 {
			out = append(out, words)
			words = nil
		}
	}

	runes := []rune(command)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case quote == '\'':
			if r == '\'' {
				quote = 0
			} else {
				word.WriteRune(r)
			}
		case quote == '"':
			switch {
			case r == '"':
				quote = 0
			case r == '\\' && i+1 < len(runes) && strings.ContainsRune("\"\\$`", runes[i+1]):
				i++
				word.WriteRune(runes[i])
			default:
				word.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote, started = r, true
		case r == '\\' && i+1 < len(runes):
			i++
			word.WriteRune(runes[i])
			started = true
		case r == ' ' || r == '\t' || r == '\r':
			endWord()
		case r == '\n' || r == ';' || r == '|' || r == '&':
			endCommand()
		default:
			word.WriteRune(r)
			started = true
		}
	}
	endCommand()
	return out
}
