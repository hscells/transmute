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
		"mp": []string{"title", "abstract", "mesh_headings"},
		"af": []string{"title", "abstract", "mesh_headings"},
		"tw": []string{"title", "abstract"},
		"nm": []string{"abstract", "mesh_headings"},
		"ab": []string{"abstract"},
		"ti": []string{"title"},
		"ot": []string{"title"},
		"sh": []string{"mesh_headings"},
		"px": []string{"mesh_headings"},
		"rs": []string{"mesh_headings"},
		"fs": []string{"mesh_headings"},
		"rn": []string{"mesh_headings"},
		"kf": []string{"mesh_headings"},
		"sb": []string{"mesh_headings"},
		"mh": []string{"mesh_headings"},
		"pt": []string{"pubtype"},

		"tiab": []string{"title", "abstract"},

		"Mesh": []string{"mesh_headings"},
		"Title": []string{"title"},
		"Title/Abstract": []string{"title", "abstract"},
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


// buildQuery takes a list of operators and keywords and constructs a boolean query. This function recursively descends
// the query group, building the immediate representation boolean query as it goes.
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

	if len(currentOp.KeywordNumbers) > 0 {
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
	} else {
		booleanQuery.Keywords = append(booleanQuery.Keywords, currentOp.Keywords...)
		for _, queryGroup := range currentOp.Children {
			booleanQuery.Children = append(booleanQuery.Children, buildQuery([]QueryGroup{queryGroup}, keywords, nil))
		}
	}

	return booleanQuery
}

// Parse a search strategy from a string of characters. There are many different ways to report a search strategy in a
// systematic review. This function attempts to determine automatically what kind of search strategy it is reading and
// parse accordingly based on some heuristics. What it cannot determine yet, however, are the character the query
// "starts after" (if there are line numbers) and what character separates a keyword or grouping from the fields being
// searched (it can be a "." or a "[").
// The output of this function is an immediate representation specified in the ir package. To compile the immediate
// representation to a search engine query, a backend must be implemented. The immediate representation is trivial to
// transform into an Elasticsearch query as it closely mirrors the Elasticsearch boolean query DSL.
func Parse(query string, startsAfter rune, fieldSeparator rune) ir.BooleanQuery {
	// TODO  this should be a map of keywordId->[]Keyword
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

		// If there is an open bracket and the field separator in the line there is a good chance the line isn't
		// a grouping of line numbers (handled below) but is instead actually an inner group of keywords.
		foundBracket := false
		foundFieldSeparator := false
		for i, char := range line {
			if char == '(' {
				foundBracket = true
			}

			if char == fieldSeparator {
				foundFieldSeparator = true
			}

			if foundBracket && i == 0 {
				foundFieldSeparator = true
			}
		}

		if foundBracket && foundFieldSeparator {
			queryGroup := parseInfixKeywords(line, startsAfter, fieldSeparator)
			queryGroup.Id = lc
			operators = append(operators, queryGroup)
			continue
		}

		// Otherwise we can parse a more typical search strategy keyword
		for i, char := range line {
			//log.Println(string(char))
			if inKeyword {
				// Now that we are definitely looking at a keyword:

				// Ignore escape characters
				if char == '\\' {
					continue
				}

				// Strip and acknowledge truncation
				if char == '*' || char == '$' {
					keyword.Truncated = true
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
					if char != startsAfter || len(keyword.QueryString) > 0 {
						keyword.QueryString += string(char)
					}

					// Check if there is an operator at the start of the string
					keywordLower := strings.ToLower(keyword.QueryString)
					if IsOperator(keywordLower) {
						queryGroup := parsePrefixGrouping(line, startsAfter)
						queryGroup.Id = lc
						operators = append(operators, queryGroup)
						isAKeyword = false
						log.Println(queryGroup)
						break
					}

					// Check if there is a number on its own and try to parse a query group
					if unicode.IsSpace(char) {
						if len(keywordLower) >= 1 && unicode.IsNumber(rune(keywordLower[0])) {
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

	log.Println(keywords)
	log.Println(operators)
	return buildQuery(operators, keywords, nil)
}