package runner

import (
	"errors"
	"strings"
)

// split cuts a command into its arguments. Single and double quotes group words and are
// dropped; nothing else is special, in particular not the backslash, which is a path separator
// on Windows.
func split(command string) ([]string, error) {
	var args []string
	var cur strings.Builder
	var quote rune
	started := false // true once an argument began, so "" still counts as one
	for _, r := range command {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote, started = r, true
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			if started {
				args = append(args, cur.String())
				cur.Reset()
				started = false
			}
		default:
			cur.WriteRune(r)
			started = true
		}
	}
	if quote != 0 {
		return nil, errors.New("the command has an unclosed quote")
	}
	if started {
		args = append(args, cur.String())
	}
	if len(args) == 0 {
		return nil, errors.New("the command is empty")
	}
	return args, nil
}
