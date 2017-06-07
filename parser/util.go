package parser

import "unicode"

// IsUpper tests to see if a string consists of all uppercase characters.
func IsUpper(s string) bool {
	for _, c := range s {
		if unicode.IsLower(c) {
			return false
		}
	}
	return true
}

// IsLower tests to see if a string consists of all lowercase characters.
func IsLower(s string) bool {
	for _, c := range s {
		if unicode.IsUpper(c) {
			return false
		}
	}
	return true
}
