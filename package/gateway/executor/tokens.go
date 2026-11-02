package executor

import (
	"errors"
	"strings"
)

func split(command string) ([]string, error) {
	var args []string
	var cur strings.Builder
	var quote rune
	started := false
	flush := func() {
		if started {
			args = append(args, cur.String())
			cur.Reset()
			started = false
		}
	}
	for _, r := range command {
		switch {
		case quote != 0 && r == quote:
			quote = 0
		case quote != 0:
			cur.WriteRune(r)
		case r == '\'' || r == '"':
			quote, started = r, true
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			flush()
		default:
			cur.WriteRune(r)
			started = true
		}
	}
	if quote != 0 {
		return nil, errors.New("the command has an unclosed quote")
	}
	flush()
	if len(args) == 0 {
		return nil, errors.New("the command is empty")
	}
	return args, nil
}
