// The grouping code handles "query groups". These are expressions in search strategies that look like:
//
// or/3-5
//
// and/6,7
//
// 9 and 10
//
// 4 and (5 or 6)
//
// etc.
//
// Generally, the type of grouping (postfix or infix) can be inferred ahead of time. This means that there needs to be
// two types of parsing methods. Additionally, only "and", "not" and "or" operators need to be taken into account.
// The prefix groups are parsed and transformed in-place, however the infix groups are parsed into a prefix tree and
// then transformed.
package parser

import (
	"strconv"
	"strings"
	"log"
	"unicode"
	"github.com/hscells/transmute/ir"
)

// QueryGroup is an intermediate structure, not part of the transmute intermediate representation, but instead used to
// construct the intermediate representation. It forms a common representation of query groupings.
type QueryGroup struct {
	// Line number of the query group. Ids of Children are 0.
	Id             int
	// Operator (one of: and, or, not)
	Type           string
	// The line numbers a keyword falls on (numbering start at 1)
	KeywordNumbers []int
	// Sometimes we just have plain old keywords
	Keywords       []ir.Keyword
	// Nested groups. for example (4 and (5 or 6)). The inner expression (5 or 6) is a child.
	Children       []QueryGroup
}

// transformPrefixGroupToQueryGroup transforms a prefix syntax tree into a query group. The new QueryGroup is built by
// recursively navigating the syntax tree.
func transformPrefixGroupToQueryGroup(prefix []string, queryGroup QueryGroup, fieldSeparator rune) ([]string, QueryGroup) {
	if len(prefix) == 0 {
		return prefix, queryGroup
	}

	token := prefix[0]
	if IsOperator(token) {
		queryGroup.Type = token
	} else if token == "(" {
		var subGroup QueryGroup
		prefix, subGroup = transformPrefixGroupToQueryGroup(prefix[1:], QueryGroup{}, fieldSeparator)
		queryGroup.Children = append(queryGroup.Children, subGroup)
	} else if token == ")" {
		return prefix, queryGroup
	} else {
		if len(token) > 0 {
			// Are we looking at a number or a keyword?
			isANumber := true
			for _, c := range token {
				if !unicode.IsNumber(c) {
					isANumber = false;
				}
			}

			// Just add the number to the list of keyword numbers
			if isANumber {
				keywordNum, err := strconv.Atoi(token)
				if err != nil {
					log.Panicln(err)
				}
				queryGroup.KeywordNumbers = append(queryGroup.KeywordNumbers, keywordNum)
			} else {
				// Otherwise we have a much more difficult string to parse
				keyword := ir.Keyword{}
				keyword.Fields = []string{}

				queryString := ""
				field := ""
				foundFieldSeparator := false
				for i, char := range token {
					// Look for the field separator
					if char == fieldSeparator {
						foundFieldSeparator = true
						continue
					} else if char == '*' || char == '$' {
						// Check if the query string is truncated
						keyword.Truncated = true
					}

					// Check if the mesh heading has been exploded
					if queryString == "exp " {
						keyword.Exploded = true
						queryString = ""
						continue
					}

					// We are in the query string state
					if !foundFieldSeparator {
						queryString += string(char)
					} else {
						// Now we are in the fields state
						if (unicode.IsPunct(char) || unicode.IsSpace(char)) && len(field) > 0 {
							keyword.Fields = append(keyword.Fields, fieldMap[field]...)
							field = ""
						} else {
							field += string(char)
						}
					}

					if i == len(token) - 1 {
						if char == '/' {
							keyword.Fields = append(keyword.Fields, "mesh_headings")
						}
					}
				}
				keyword.QueryString = queryString
				queryGroup.Keywords = append(queryGroup.Keywords, keyword)
			}
		}
	}
	return transformPrefixGroupToQueryGroup(prefix[1:], queryGroup, fieldSeparator)
}

// convertInfixToPrefix translates an infix grouping expression into a prefix expression. The way this is done is the
// Shunting-yard algorithm (https://en.wikipedia.org/wiki/Shunting-yard_algorithm).
func convertInfixToPrefix(infix []string) []string {
	// The stack contains some intermediate values
	stack := []string{}
	// The result contains the actual expression
	result := []string{}

	precedence := map[string]int{
		"and": 1,
		"or": 0,
		"not": 1,
		"adj": 1,
		"adj2": 1,
		"adj3": 1,
		"adj4": 1,
		"adj5": 1,
		"adj6": 1,
		"adj7": 1,
		"adj8": 1,
	}

	// The algorithm is slightly modified to also store the brackets in the result
	for i := len(infix) - 1; i >= 0; i-- {
		token := infix[i]
		if token == ")" {
			stack = append(stack, token)
			result = append(result, token)
		} else if token == "(" {
			for len(stack) > 0 {
				var t string
				t, stack = stack[len(stack) - 1], stack[:len(stack) - 1]
				if t == ")" {
					result = append(result, "(")
					break
				}
				result = append(result, t)
			}
		} else if _, ok := precedence[token]; !ok {
			result = append(result, token)
		} else {
			for len(stack) > 0 && precedence[stack[len(stack) - 1]] > precedence[token] {
				var t string
				t, stack = stack[len(stack) - 1], stack[:len(stack) - 1]
				result = append(result, t)
			}
			stack = append(stack, token)
		}

	}

	for len(stack) > 0 {
		var t string
		t, stack = stack[len(stack) - 1], stack[:len(stack) - 1]
		result = append(result, t)
	}

	// The algorithm actually produces a postfix expression so it must be reversed
	// We can do this in-place with go!
	for i := len(result) / 2 - 1; i >= 0; i-- {
		opp := len(result) - 1 - i
		result[i], result[opp] = result[opp], result[i]
	}

	return result
}

func addFieldsToKeywords(queryGroup QueryGroup, fields []string) QueryGroup {
	for i, keyword := range queryGroup.Keywords {
		queryGroup.Keywords[i].Fields = append(keyword.Fields, fields...)
	}

	for _, child := range queryGroup.Children {
		child = addFieldsToKeywords(child, fields)
	}

	return queryGroup
}

func parseInfixKeywords(line string, startsAfter, fieldSeparator rune) QueryGroup {
	line += "\n"

	stack := []string{}
	inGroup := startsAfter == rune(0)

	keyword := ""
	currentToken := ""
	previousToken := ""

	endTokens := ""

	depth := 0

	for _, char := range line {
		// Ignore the first few characters of the line
		if !inGroup && char == startsAfter {
			inGroup = true
			continue
		} else if !inGroup {
			continue
		}

		if unicode.IsSpace(char) {
			t := strings.ToLower(currentToken)
			if IsOperator(t) {
				keyword = previousToken
				stack = append(stack, strings.TrimSpace(keyword))
				stack = append(stack, strings.TrimSpace(t))
				previousToken = ""
				keyword = ""
			} else {
				previousToken += " " + currentToken
			}
			currentToken = ""
			continue
		} else if char == '(' {
			depth++
			stack = append(stack, "(")
			currentToken = ""
			continue
		} else if char == ')' {
			depth--
			if len(keyword) > 0 || len(currentToken) > 0 {
				stack = append(stack, strings.TrimSpace(keyword + " " + currentToken))
				keyword = ""
				currentToken = ""
			}
			stack = append(stack, ")")
			continue
		} else if !unicode.IsSpace(char) {
			currentToken += string(char)
		}


		if depth == 0 {
			endTokens += string(char)
		}
	}

	stack = append(stack, endTokens)
	// The end of the expression contains fields, so we need to populate the fields of all the keywords.
	fields := []string{}
	if stack[len(stack) - 1] != ")" {
		field := ""
		for _, char := range(stack[len(stack) - 1]) {
			if char == fieldSeparator {
				continue
			} else if (unicode.IsPunct(char) || unicode.IsSpace(char)) && len(field) > 0 {
				fields = append(fields, fieldMap[field]...)
				field = ""
			} else {
				field += string(char)
			}
		}
		fields = append(fields, fieldMap[field]...)

		stack = stack[:len(stack) - 2]
	}
	prefix := convertInfixToPrefix(stack)
	_, queryGroup := transformPrefixGroupToQueryGroup(prefix, QueryGroup{}, fieldSeparator)

	if len(fields) > 0 {
		queryGroup = addFieldsToKeywords(queryGroup, fields)
	}

	return queryGroup
}

// parseInfixGrouping translates an infix grouping into a prefix grouping, and then transforms it into a QueryGroup in
// two separate steps.
func parseInfixGrouping(group string, startsAfter rune) QueryGroup {
	group += "\n"
	group = strings.ToLower(group)

	stack := []string{}
	inGroup := startsAfter == rune(0)
	keyword := ""

	for _, char := range group {
		// Ignore the first few characters of the line
		if !inGroup && char == startsAfter {
			inGroup = true
			continue
		} else if !inGroup {
			continue
		}

		if unicode.IsSpace(char) && len(keyword) > 0 {
			stack = append(stack, strings.TrimSpace(keyword))
			keyword = ""
			continue
		} else if char == '(' {
			stack = append(stack, "(")
			keyword = ""
			continue
		} else if char == ')' {
			if len(keyword) > 0 {
				stack = append(stack, strings.TrimSpace(keyword))
				keyword = ""
			}
			stack = append(stack, ")")
			continue
		} else if !unicode.IsSpace(char) {
			keyword += string(char)
		}
	}

	prefix := convertInfixToPrefix(stack)
	_, queryGroup := transformPrefixGroupToQueryGroup(prefix, QueryGroup{}, 0)
	return queryGroup
}

// parseGrouping parses and constructs a QueryGroup in-place. Since the grouping is post-fix, no additional
// transformation is necessary.
func parsePrefixGrouping(group string, startsAfter rune) QueryGroup {
	group += "\n"

	var nums []string
	var num string

	var sep string

	var operator string

	queryGroup := QueryGroup{}

	inGroup := false
	if startsAfter == rune(0) {
		inGroup = true
	}

	for _, char := range group {
		// Set the separator
		if char == '-' {
			sep = "-"
		} else if char == ',' {
			sep = ","
		}

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

			// Now, unfortunately there is an infix operator INSIDE the postfix expression.
			// This is parsed as follows:
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

			} else if len(nums) == 1 {
				if sep == "," {
					lhs, err := strconv.Atoi(nums[0])
					if err != nil {
						panic(err)
					}
					queryGroup.Type = operator
					queryGroup.KeywordNumbers = append(queryGroup.KeywordNumbers, lhs)
					nums = []string{}
				}
			}

		}

		// Extract the groups
		if IsOperator(operator) {
			operator += strings.ToLower(string(char))
		}

	}
	return queryGroup
}
