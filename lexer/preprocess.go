package lexer

import (
	"strings"
	"unicode"
)

func PreProcess(query string) string {
	l := strings.TrimSpace(strings.Split(query, "\n")[0])
	if !strings.Contains(l, " ") {
		return query
	} else if !strings.ContainsAny(strings.Split(l, " ")[0], "0123456789") {
		return query
	}

	newQuery := ""
	for _, line := range strings.Split(query, "\n") {
		queryString := ""
		foundStart := false
		for _, char := range line {
			if !foundStart && (unicode.IsSymbol(char) || unicode.IsNumber(char) ||
				char == '#' || char == '.') {
				continue
			}

			if !foundStart && unicode.IsSpace(char) {
				foundStart = true
				continue
			}

			if foundStart {
				queryString += string(char)
			}
		}
		queryString = strings.Replace(queryString, "\\", "", -1)
		queryString = strings.TrimSpace(queryString)
		newQuery += strings.ToLower(queryString) + "\n"
	}
	return newQuery
}
