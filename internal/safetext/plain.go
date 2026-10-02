package safetext

import "strings"

// Plain removes terminal control characters from untrusted single-line text.
// Apply it before adding application-owned ANSI styling.
func Plain(value string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			return ' '
		}
		return r
	}, value)
}
