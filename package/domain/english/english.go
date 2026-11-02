package english

import (
	"strings"
	"unicode"
)

const Code = "LANG-NOT-ENGLISH"

const accented = "áéíóúñüàèìòùâêîôûçãõäëïöß¿¡"

var spanish = map[string]bool{
	"el": true, "la": true, "los": true, "las": true, "del": true, "que": true, "una": true, "por": true,
	"para": true, "con": true, "se": true, "su": true, "al": true, "como": true, "pero": true,
	"cuando": true, "debe": true, "deben": true, "puede": true, "cada": true, "sin": true, "sobre": true,
	"entre": true, "también": true, "siempre": true, "nunca": true, "porque": true, "donde": true,
}

func NotEnglish(text string) bool {
	text = strings.ToLower(text)
	words, foreign, accents := 0, 0, 0
	for _, w := range strings.FieldsFunc(text, isSeparator) {
		words++
		if spanish[w] {
			foreign++
		}
	}
	for _, r := range text {
		if strings.ContainsRune(accented, r) {
			accents++
		}
	}
	return accents >= 4 || (foreign >= 4 && foreign*12 >= words)
}

func isSeparator(r rune) bool { return !unicode.IsLetter(r) && r != '\'' }
