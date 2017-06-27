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

func IsOperator(s string) bool {
	return s == "or" ||
		s == "and" ||
		s == "not" ||
		s == "adj" ||
		s == "adj2" ||
		s == "adj3" ||
		s == "adj4" ||
		s == "adj5" ||
		s == "adj6" ||
		s == "adj7" ||
		s == "adj8"
}