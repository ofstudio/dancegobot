package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSingleNumber(t *testing.T) {
	h := &Handlers{}

	tests := []struct {
		text   string
		want   int
		wantOK bool
	}{
		{text: "1", want: 1, wantOK: true},
		{text: "1.", want: 1, wantOK: true},
		{text: "1. Alice", want: 1, wantOK: true},
		{text: "01. Alice", want: 1, wantOK: true},
		{text: "1 Alice", wantOK: false},
		{text: "Alice 1.", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got, gotOK := h.singleNumber(tt.text)
			require.Equal(t, tt.wantOK, gotOK)
			require.Equal(t, tt.want, got)
		})
	}
}
