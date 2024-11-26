package numerals

// NumeralRU is a formatter for nouns after numerals in Russian.
// It is used to format the noun in the correct form depending on the numeral.
//
// The first element is the nominative form, the second is the singular genitive form,
// and the third is the plural genitive form.
//
// Example: "день", "дня", "дней"
type NumeralRU [3]string

// Ru returns a new formatter for nouns after numerals in Russian.
// Arguments are the nominative, singular genitive and plural genitive forms of the noun.
// Example:
//
//	Ru("день", "дня", "дней").N(2) // returns "дня"
func Ru(n, sg, pg string) NumeralRU {
	return NumeralRU{n, sg, pg}
}

// N returns the noun in the correct form depending on the numeral.
func (n NumeralRU) N(num int) string {
	rem := num % 100
	if rem < 0 {
		rem = -rem
	}
	if rem > 19 {
		rem %= 10
	}
	switch rem {
	case 1:
		return n[0]
	case 2, 3, 4:
		return n[1]
	default:
		return n[2]
	}
}
