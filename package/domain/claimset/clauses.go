package claimset

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var leaders = map[string]bool{
	"when": false, "cuando": false,
	"while": false, "mientras": false,
	"where": false, "donde": false,
	"if": true, "si": true,
}

var thens = map[string]bool{"then": true, "entonces": true}

func checkEARS(text string) error {
	rest := text
	for {
		word, after := cutWord(rest)
		needsThen, ok := leaders[strings.ToLower(word)]
		if !ok {
			break
		}
		clause, tail, found := strings.Cut(after, ",")
		if !found {
			return fmt.Errorf("the %q clause needs a comma before the response", word)
		}
		if strings.TrimSpace(clause) == "" {
			return fmt.Errorf("the %q clause is empty", word)
		}
		rest = strings.TrimSpace(tail)
		if needsThen {
			w, after := cutWord(rest)
			if !thens[strings.ToLower(w)] {
				return fmt.Errorf("the %q clause needs \"then\" after the comma", word)
			}
			rest = after
		}
	}
	if rest == "" {
		return errors.New("the sentence has no response")
	}
	return nil
}

func cutWord(s string) (word, rest string) {
	i := strings.IndexFunc(s, unicode.IsSpace)
	if i < 0 {
		return s, ""
	}
	return s[:i], strings.TrimSpace(s[i:])
}
