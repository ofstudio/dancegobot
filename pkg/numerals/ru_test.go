package numerals

import "testing"

func TestNumeralRU_N(t *testing.T) {
	tests := []struct {
		name     string
		num      int
		expected string
	}{
		{"1 day", 1, "день"},
		{"2 days", 2, "дня"},
		{"5 days", 5, "дней"},
		{"21 days", 21, "день"},
		{"22 days", 22, "дня"},
		{"25 days", 25, "дней"},
		{"0 days", 0, "дней"},
		{"1234123 days", 1234123, "дня"},
		{"negative 1 day", -1, "день"},
		{"negative 2 days", -2, "дня"},
		{"negative 5 days", -5, "дней"},
	}

	numeral := Ru("день", "дня", "дней")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := numeral.N(tt.num); got != tt.expected {
				t.Errorf("NumeralRU.N() = %v, want %v", got, tt.expected)
			}
		})
	}
}
