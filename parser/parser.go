// Package parser implements a parser for the search strategies in systematic reviews. The goal of the parser is to
// transform it into an immediate representation that can then be translated into queries suitable for other systems.
package parser

import (
	"io/ioutil"
	"github.com/hscells/transmute/ir"
	"strings"
	"unicode"
	"strconv"
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

type QueryGroup struct {
	Type           string
	KeywordNumbers []int
	Children       []QueryGroup
}

// Load a search strategy from a file
func Load(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// parseGrouping parses and constructs a QueryGroup.
func parseGrouping(group string, startsAfter rune) QueryGroup {
	group += "\n"

	var nums []string
	var num string

	var sep string

	var operator string

	queryGroup := QueryGroup{}

	inGroup := false
	for _, char := range group {
		// Ignore the first few characters of the line
		if !inGroup && char == startsAfter {
			inGroup = true
			continue
		} else if !inGroup {
			continue
		}

		// Extract the numbers
		if unicode.IsNumber(char) {
			num += string(char)
		} else if len(num) > 0 {
			nums = append(nums, num)
			num = ""

			if len(nums) == 2 {
				if sep == "-" {
					queryGroup.Type = operator

					lhs, err := strconv.Atoi(nums[0])
					if err != nil {
						panic(err)
					}

					rhs, err := strconv.Atoi(nums[1])
					if err != nil {
						panic(err)
					}

					for i := lhs; i <= rhs; i++ {
						queryGroup.KeywordNumbers = append(queryGroup.KeywordNumbers, i)
					}

				}
				nums = []string{}

			}

		}


		// Extract the groups
		if operator != "or" && operator != "and" {
			operator += strings.ToLower(string(char))
		}

		// Set the separator
		if char == '-' {
			sep = "-"
		}

	}
	return queryGroup
}

// buildQuery takes a list of operators and keywords and constructs a boolean query.
func buildQuery(operators []QueryGroup, keywords []ir.Keyword) ir.BooleanQuery {
	booleanQuery := ir.BooleanQuery{}

	for i := len(operators) - 1; i >= 0; i-- {
		booleanQuery.Operator = operators[i].Type

		for k := range operators[i].KeywordNumbers {
			booleanQuery.Keywords = append(booleanQuery.Keywords, keywords[k])
		}

		if len(operators[i].Children) > 0 {
			booleanQuery.Children = append(booleanQuery.Children, buildQuery(operators[i].Children, keywords))
		} else {
			booleanQuery.Children = []ir.BooleanQuery{}
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

		keyword := ir.Keyword{Fields: make([]string,0)}

		inKeyword := false
		isAKeyword := true
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
						operators = append(operators, parseGrouping(line, startsAfter))
						isAKeyword = false
						break
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
			keywords = append(keywords, keyword)
		}
	}

	return buildQuery(operators, keywords)
}