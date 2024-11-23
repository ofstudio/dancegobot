package helpers

import (
	"strings"
)

// Errors represents a map of errors.
type Errors map[string]error

// Filter removes nil errors from the map.
// Returns nil if there are no errors left.
func (e Errors) Filter() error {
	for key, val := range e {
		if val == nil {
			delete(e, key)
		}
	}
	if len(e) == 0 {
		return nil
	}
	return e
}

// Error returns the error string of Errors.
func (e Errors) Error() string {
	sb := strings.Builder{}
	i := 0
	for key, val := range e {
		if val == nil {
			continue
		}
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString(val.Error())
		i++
	}
	return sb.String()
}
