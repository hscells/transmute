// Package parser implements a parser for the search strategies in systematic reviews. The goal of the parser is to
// transform it into an immediate representation that can then be translated into queries suitable for other systems.
package parser

import (
	"io/ioutil"
	"github.com/hscells/transmute/ir"
	"strings"
	"unicode"
	"log"
)

var (
	fieldMap = map[string][]string{
		"mp": []string{"title", "text", "mesh_headings"},
		"af": []string{"title", "text", "mesh_headings"},
		"tw": []string{"title", "text"},
		"nm": []string{"text", "mesh_headings"},
		"ab": []string{"text"},
		"ti": []string{"title"},
		"ot": []string{"title"},
		"sh": []string{"mesh_headings"},
		"px": []string{"mesh_headings"},
		"rs": []string{"mesh_headings"},
		"fs": []string{"mesh_headings"},
		"rn": []string{"mesh_headings"},
		"kf": []string{"mesh_headings"},
		"sb": []string{"mesh_headings"},
		"pt": []string{"pubtype"},
	}
)

// Load a search strategy from a file.
func Load(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(data)
}


// buildQuery takes a list of operators and keywords and constructs a boolean query.
func buildQuery(operators []QueryGroup, keywords []ir.Keyword, seenIds []int) ir.BooleanQuery {
	booleanQuery := ir.BooleanQuery{}

	if seenIds == nil {
		seenIds = make([]int, 0)
	}

	currentOp := operators[len(operators) - 1]

	//log.Println(currentOp)

	for _, id := range seenIds {
		if currentOp.Id == id {
			return booleanQuery
		}
	}

	seenIds = append(seenIds, currentOp.Id)

	booleanQuery.Operator = currentOp.Type
	booleanQuery.Children = []ir.BooleanQuery{}
	booleanQuery.Keywords = []ir.Keyword{}

	for _, keywordId := range currentOp.KeywordNumbers {
		for _, j := range operators {
			if j.Id == keywordId {
				booleanQuery.Children = append(booleanQuery.Children, buildQuery(append(operators, j), keywords, seenIds))
			}
		}

		for _, keyword := range keywords {
			if keyword.Id == keywordId {
				booleanQuery.Keywords = append(booleanQuery.Keywords, keyword)
			}
		}

		if len(currentOp.Children) > 0 {
			booleanQuery.Children = append(booleanQuery.Children, buildQuery(currentOp.Children, keywords, nil))
		}
	}

	return booleanQuery
}

// Parse a search strategy from a string of characters
func Parse(query string, startsAfter rune, fieldSeparator rune) ir.BooleanQuery {
	keywords := []ir.Keyword{}
	operators := []QueryGroup{}

	// Line count / line data
	for lc, line := range strings.Split(query, "\n") {
		lc = lc + 1

		keyword := ir.Keyword{Fields: make([]string, 0)}
		keyword.Id = lc

		isAKeyword := true
		inKeyword := startsAfter == rune(0)
		seenFieldSep := false
		currentField := ""

		for i, char := range line {
			if inKeyword {
				// Now that we are definitely looking at a keyword:

				// Ignore escape characters
				if char == '\\' {
					continue
				}

				// Strip and acknowledge truncation
				if char == '*' {
					keyword.Truncated = true
					continue
				}

				// Look for an `exploded` mesh heading
				if keyword.QueryString == "exp" {
					keyword.Exploded = true
					keyword.QueryString = ""
					continue
				}

				// Look for mesh heading line terminator
				if i == len(line) - 1 && char == '/' {
					keyword.Fields = append(keyword.Fields, "mesh_headings")
					break
				}

				if !seenFieldSep && char == fieldSeparator {
					seenFieldSep = true
					continue
				}

				if !seenFieldSep {
					// keywords

					// Continue building the query string
					keyword.QueryString += string(char)

					// Check if there is an operator at the start of the string
					keywordLower := strings.ToLower(keyword.QueryString)
					if keywordLower == "and" || keywordLower == "or" || keywordLower == "not" {
						queryGroup := parsePrefixGrouping(line, startsAfter)
						queryGroup.Id = lc
						operators = append(operators, queryGroup)
						isAKeyword = false
						break
					}

					// Check if there is a number on its own and try to parse a query group
					if unicode.IsSpace(char) {
						isNumber := true;
						for i, c := range keywordLower {
							if i < len(keywordLower) - 1 && !unicode.IsNumber(c) {
								isNumber = false
								break
							}
						}

						if isNumber {
							queryGroup := parseInfixGrouping(line, startsAfter)
							queryGroup.Id = lc
							operators = append(operators, queryGroup)
							isAKeyword = false
							break
						}

					}

				} else if seenFieldSep {
					// fields
					if unicode.IsPunct(char) {
						if fields, ok := fieldMap[currentField]; ok {
							keyword.Fields = append(keyword.Fields, fields...)
						} else {
							log.Panicf("Cannot find mapping for field %v", currentField)
						}

						currentField = ""
					} else {
						currentField += string(char)
					}
				}

			} else if !inKeyword && char != startsAfter {
				// This is not a keyword and is still the start of the line
				continue
			} else {
				// We got to a state that is the start of the keyword
				inKeyword = true
			}
		}

		// Ignore lines where we are looking at for instance, an operator
		if isAKeyword {
			if len(keyword.Fields) == 0 {
				if IsUpper(keyword.QueryString) {
					keyword.Fields = append(keyword.Fields, "mesh_headings")
				} else {
					keyword.Fields = append(keyword.Fields, "title", "text")
				}
			}

			keywords = append(keywords, keyword)
		}
	}

	return buildQuery(operators, keywords, nil)
}