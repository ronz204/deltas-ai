// Package english tells whether the text of the knowledge base is written in English, which is the only
// language it is written in, whatever language the person speaks to the agent.
package english

import (
	"strings"
	"unicode"
)

// Code is the diagnostic code of a text that is not in English. Hooks and CI depend on it, so it does not change.
const Code = "LANG-NOT-ENGLISH"

// spanish holds common Spanish words that are not English words. It is a heuristic for the language the
// agent is most likely to slip into, not a language detector.
var spanish = map[string]bool{
	"el": true, "la": true, "los": true, "las": true, "del": true, "que": true, "una": true, "por": true,
	"para": true, "con": true, "se": true, "su": true, "al": true, "como": true, "pero": true,
	"cuando": true, "debe": true, "deben": true, "puede": true, "cada": true, "sin": true, "sobre": true,
	"entre": true, "también": true, "siempre": true, "nunca": true, "porque": true, "donde": true,
}

// accented are letters that English text almost never carries; a few of them in a spec are a sign of another language.
const accented = "áéíóúñüàèìòùâêîôûçãõäëïöß¿¡"

// NotEnglish says whether text is clearly not English: several accented letters, or a clear share of Spanish
// words. A name or a quoted word in another language does not trip it.
func NotEnglish(text string) bool {
	words, foreign, letters := 0, 0, 0
	for _, w := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool { return !unicode.IsLetter(r) && r != '\'' }) {
		words++
		if spanish[w] {
			foreign++
		}
	}
	for _, r := range strings.ToLower(text) {
		if strings.ContainsRune(accented, r) {
			letters++
		}
	}
	return letters >= 4 || (foreign >= 4 && foreign*12 >= words)
}
