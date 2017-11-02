package parser

import (
	"github.com/hscells/transmute/ir"
	"strings"
	"unicode"
	"log"
)

type PubMedTransformer struct{}

var PubMedFieldMapping = map[string][]string{
	"Mesh":           {"mesh_headings"},
	"mesh":           {"mesh_headings"},
	"MeSH":           {"mesh_headings"},
	"Title/Abstract": {"title", "abstract"},
	"Title":          {"title"},
	"Abstract":       {"abstract"},
	"Publication":    {"pub_type"},
	"mh":             {"mesh_headings"},
	"sh":             {"mesh_headings"},
	"tw":             {"title", "abstract"},
	"ti":             {"title"},
	"pt":             {"pub_type"},
	"sb":             {"pub_status"},
	"tiab":           {"title", "abstract"},
	"default":        {"abstract"},
}

func (t PubMedTransformer) TransformSingle(query string, mapping map[string][]string) ir.Keyword {
	var queryString string
	var fields []string
	exploded := true

	if strings.ContainsRune(query, '[') {
		// This query string most likely has a field.
		parts := strings.Split(query, "[")
		queryString = parts[0]
		// This might be a field, but needs some processing.
		possibleField := strings.Replace(parts[1], "]", "", -1)

		// PubMed fields have this weird thing where they specify the mesh explosion in the field.
		// This is handled in this step.
		if strings.Contains(strings.ToLower(possibleField), ":noexp") {
			exploded = false
			possibleField = strings.Replace(strings.ToLower(possibleField), ":noexp", "", -1)
		}

		// If we are unable to map the field then we can explode.
		if field, ok := mapping[possibleField]; ok {
			fields = field
		} else {
			log.Fatalf("the field `%v` does not have a mapping defined", possibleField)
		}
	} else {
		queryString = query
	}

	// Add a default field to the keyword if none have been defined
	if len(fields) == 0 {
		fields = mapping["default"]
	}

	// medline uses $ to represent the stem of a word. Instead let's just replace it by the wildcard operator.
	// TODO is there anything in Elasticsearch to do this? - and by `this` I mean single character wildcards.
	queryString = strings.Replace(queryString, "$", "*", -1)

	return ir.Keyword{
		QueryString: queryString,
		Fields:      fields,
		Exploded:    exploded,
		Truncated:   false,
	}
}

func (t PubMedTransformer) TransformNested(query string, mapping map[string][]string) ir.BooleanQuery {
	var fieldsString string
	for i := len(query) - 1; i > 0; i-- {
		if query[i] == ')' {
			break
		}
		fieldsString += string(query[i])
	}
	fieldsString = ReversePreservingCombiningCharacters(fieldsString)
	query = strings.Replace(query, fieldsString, "", 1)

	return t.ParseInfixKeywords(query, mapping)
}

// ParseInfixKeywords parses an infix expression containing keywords separated by operators into an infix expression,
// and then into the immediate representation.
func (t PubMedTransformer) ParseInfixKeywords(line string, mapping map[string][]string) ir.BooleanQuery {
	line += "\n"

	stack := []string{}

	keyword := ""
	currentToken := ""
	previousToken := ""

	endTokens := ""

	depth := 0
	insideQuote := false

	for _, char := range line {
		// Here we attempt to parse a keyword that is quoted.
		if char == '"' && !insideQuote {
			insideQuote = true
			continue
		} else if char == '"' && insideQuote {
			insideQuote = false
			continue
		} else if insideQuote {
			currentToken += string(char)
			continue
		}

		if unicode.IsSpace(char) {
			tok := strings.ToLower(currentToken)
			if t.IsOperator(tok) {
				keyword = previousToken
				stack = append(stack, strings.TrimSpace(keyword))
				stack = append(stack, strings.TrimSpace(tok))
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
				stack = append(stack, strings.TrimSpace(keyword+" "+currentToken))
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

	if len(endTokens) > 0 {
		stack = append(stack, endTokens)
	}
	prefix := t.ConvertInfixToPrefix(stack)
	if prefix[0] == "(" && prefix[len(prefix)-1] == ")" {
		prefix = prefix[1:len(prefix)-1]
	}
	log.Println(prefix)
	_, queryGroup := t.TransformPrefixGroupToQueryGroup(prefix, ir.BooleanQuery{}, mapping)
	return queryGroup
}

// ConvertInfixToPrefix translates an infix grouping expression into a prefix expression. The way this is done is the
// Shunting-yard algorithm (https://en.wikipedia.org/wiki/Shunting-yard_algorithm).
func (t PubMedTransformer) ConvertInfixToPrefix(infix []string) []string {
	// The stack contains some intermediate values
	stack := []string{}
	// The result contains the actual expression
	result := []string{}

	precedence := map[string]int{
		"and":  1,
		"or":   0,
		"not":  1,
		"adj":  1,
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
				t, stack = stack[len(stack)-1], stack[:len(stack)-1]
				if t == ")" {
					result = append(result, "(")
					break
				}
				result = append(result, t)
			}
		} else if _, ok := precedence[token]; !ok {
			result = append(result, token)
		} else {
			for len(stack) > 0 && precedence[stack[len(stack)-1]] > precedence[token] {
				var t string
				t, stack = stack[len(stack)-1], stack[:len(stack)-1]
				result = append(result, t)
			}
			stack = append(stack, token)
		}

	}

	for len(stack) > 0 {
		var t string
		t, stack = stack[len(stack)-1], stack[:len(stack)-1]
		result = append(result, t)
	}

	// The algorithm actually produces a postfix expression so it must be reversed
	// We can do this in-place with go!
	for i := len(result)/2 - 1; i >= 0; i-- {
		opp := len(result) - 1 - i
		result[i], result[opp] = result[opp], result[i]
	}

	return result
}

// IsOperator tests to see if a string is a valid PubMed/Medline operator.
func (t PubMedTransformer) IsOperator(s string) bool {
	return s == "or" ||
		s == "and" ||
		s == "not" ||
		adjMatchRegexp.MatchString(s)
}

// transformPrefixGroupToQueryGroup transforms a prefix syntax tree into a query group. The new QueryGroup is built by
// recursively navigating the syntax tree.
func (t PubMedTransformer) TransformPrefixGroupToQueryGroup(prefix []string, queryGroup ir.BooleanQuery, mapping map[string][]string) ([]string, ir.BooleanQuery) {
	if len(prefix) == 0 {
		return prefix, queryGroup
	}

	token := prefix[0]
	if t.IsOperator(token) {
		queryGroup.Operator = token
	} else if token == "(" {
		var subGroup ir.BooleanQuery
		prefix, subGroup = t.TransformPrefixGroupToQueryGroup(prefix[1:], ir.BooleanQuery{}, mapping)
		if len(subGroup.Operator) == 0 {
			if len(queryGroup.Keywords) > 0 {
				queryGroup.Keywords = append(queryGroup.Keywords, subGroup.Keywords...)
			} else {
				queryGroup.Keywords = subGroup.Keywords
			}
		} else {
			queryGroup.Children = append(queryGroup.Children, subGroup)
		}
	} else if token == ")" {
		return prefix, queryGroup
	} else {
		if len(token) > 0 {
			k := t.TransformSingle(token, mapping)
			queryGroup.Keywords = append(queryGroup.Keywords, k)
		}
	}
	return t.TransformPrefixGroupToQueryGroup(prefix[1:], queryGroup, mapping)
}

func NewPubMedParser() QueryParser {
	return QueryParser{FieldMapping: PubMedFieldMapping, Parser: PubMedTransformer{}}
}
