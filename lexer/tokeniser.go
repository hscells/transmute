package lexer

import "regexp"

// A tokeniser is used to read characters from a text stream and match them to patterns to produce tokens.
type Tokeniser struct {
	// The current token.
	token string

	// Compiled patterns for the lexer to match.
	patterns map[string]regexp.Regexp
}

// A token is produced from the tokeniser when a Valid token has been seen by the lexer.
type Token struct {
	// The string representation of the token.
	Value string

	// What Pattern the lexer matched to output a token.
	Pattern regexp.Regexp

	// What is the pattern referencing in the ast?
	Reference string

	// Is the token Valid?
	Valid bool
}


// NewTokeniser creates a tokeniser from one or more regex strings.
func NewTokeniser(patterns map[string]string) (Tokeniser, error) {
	compiledPatterns := map[string]regexp.Regexp{}

	for reference, pattern := range patterns {
		compiledPattern, err := regexp.Compile(pattern)
		if err != nil {
			return Tokeniser{}, err
		}

		compiledPatterns[reference] = *compiledPattern
	}


	return Tokeniser{
		patterns: compiledPatterns,
	}, nil
}

// Consume a character and output tokens.
func (t *Tokeniser) Consume(char rune) Token {
	t.token += string(char)
	for reference, pattern := range t.patterns {
		if pattern.MatchString(t.token) {
			//println(reference)
			return Token{
				Value: t.token,
				Pattern: pattern,
				Reference: reference,
				Valid:   true,
			}
		}
	}
	return Token{Valid: false}
}
