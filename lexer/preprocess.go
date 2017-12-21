package lexer

import (
	"strings"
	"unicode"
)

// PreProcess attempts to remove the starting numbers from a query and will trim each line in a query if there are any
// additional, unnecessary spaces. The output should be a fairly clean search strategy.
func PreProcess(query string, options LexOptions) string {
	// Format the parenthesis
	if options.FormatParenthesis {
		query = strings.Replace(query, ")", " ) ", -1)
		query = strings.Replace(query, "(", " ( ", -1)
	}

	// Identify queries as single line queries or search strategies without numbers.
	l := strings.TrimSpace(strings.Split(query, "\n")[0])
	if !strings.Contains(l, " ") {
		return query
	} else if !strings.ContainsAny(strings.Split(l, " ")[0], "0123456789") {
		return query
	}

	// Otherwise just process each line at a time.
	newQuery := ""
	for _, line := range strings.Split(query, "\n") {
		line = strings.TrimSpace(line)
		queryString := ""
		foundStart := false
		for _, char := range line {
			// Skip if it's not a valid query string character.
			if !foundStart && (unicode.IsSymbol(char) || unicode.IsNumber(char) ||
				char == '#' || char == '.') {
				continue
			}

			// Skip if it's a space and there is no start.
			if !foundStart && unicode.IsSpace(char) {
				foundStart = true
				continue
			}

			// Now we have probably found the query string.
			if foundStart {
				queryString += string(char)
			}
		}

		// Format the query string.
		queryString = strings.Replace(queryString, "\\", "", -1)
		queryString = strings.TrimSpace(queryString)
		newQuery += queryString + "\n"
	}
	return newQuery
}
