package claimset

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var scenario = regexp.MustCompile(`(?i)^(?:given|dad[oa](?: que)?)\s+(.+?),\s*(?:when|cuando)\s+(.+?),\s*(?:then|entonces)\s+(.+)$`)

func CheckStatement(k Kind, text string) error {
	text = strings.TrimSpace(text)
	switch k {
	case Invariant:
		return checkEARS(text)
	case Acceptance:
		if !scenario.MatchString(text) {
			return errors.New(`an acceptance criterion reads "Given <context>, when <action>, then <outcome>"`)
		}
		return nil
	}
	return fmt.Errorf("a claim of kind %q has no sentence notation", k)
}
