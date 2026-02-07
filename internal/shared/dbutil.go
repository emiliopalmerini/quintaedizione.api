package shared

import "strings"

// EscapeLike escapes special characters (%, _, \) in a string
// intended for use in SQL LIKE/ILIKE patterns.
func EscapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
