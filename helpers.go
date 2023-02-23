package badcli

import "strings"
import "unicode/utf8"

func runeLenAbove(str string, n int) bool {
	// fast cases
	if len(str)   <= n { return false }
	if len(str)/4 >= n { return true  }

	// general case
	index := 0
	for n >= 0 && index < len(str) {
		_, runeSize := utf8.DecodeRuneInString(str[index : ])
		index += runeSize
		n -= 1
	}
	return n < 0
}

func hasAnySuffix(str string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(str, suffix) {
			return true
		}
	}
	return false
}
