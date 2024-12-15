package services

import (
	"regexp"
	"strings"
	"time"
)

// nowFn returns the current time in UTC.
var nowFn = func() time.Time {
	return time.Now().UTC()
}

// reUsername is a regular expression to extract telegram username from the string.
// Regex explanation:
//
//	(boundary)@(username)(boundary)
var reUsername = regexp.MustCompile(`(?:^|\b|\s)@([a-zA-Z][a-zA-Z0-9_]{3,30}[a-zA-Z0-9])(?:\b|$)`)

// username extracts Telegram @username from string.
// Only the first username occurrence will be extracted.
//
// Telegram username rules:
//   - 5-32 characters (without @)
//   - can contain only letters, digits and underscores: a-z, A-Z, 0-9, _
//   - must start with a letter
//   - must end with a letter or a digit
func username(s string) (string, bool) {
	matches := reUsername.FindStringSubmatch(s)
	if len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}

// errMap represents a map of errors.
// It implements the error interface and used to collect multiple errors.
type errMap map[string]error

// Filter removes nil errors from the map.
// Returns nil if there are no errors left.
func (e errMap) Filter() error {
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

// Error returns the error string of errMap.
func (e errMap) Error() string {
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
