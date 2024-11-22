package helpers

import "regexp"

// reUsername is a regular expression to extract telegram username from the string.
// Regex explanation:
//
//	(boundary)@(username)(boundary)
var reUsername = regexp.MustCompile(`(?:^|\b|\s)@([a-zA-Z][a-zA-Z0-9_]{3,30}[a-zA-Z0-9])(?:\b|$)`)

// Username extracts Telegram @username from string.
// Only the first username occurrence will be extracted.
//
// Telegram username rules:
//   - 5-32 characters (without @)
//   - can contain only letters, digits and underscores: a-z, A-Z, 0-9, _
//   - must start with a letter
//   - must end with a letter or a digit
func Username(s string) (string, bool) {
	matches := reUsername.FindStringSubmatch(s)
	if len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}
