package parser

import (
	"github.com/hscells/transmute/ir"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
	"log"
)

var MedlineFieldMapping = map[string][]string{
	"mp":      {"mesh_headings"},
	"af":      {"title", "abstract", "mesh_headings"},
	"tw":      {"title", "abstract"},
	"nm":      {"abstract", "mesh_headings"},
	"ab":      {"abstract"},
	"ti":      {"title"},
	"ot":      {"title"},
	"sh":      {"mesh_headings"},
	"px":      {"mesh_headings"},
	"rs":      {"mesh_headings"},
	"fs":      {"mesh_headings"},
	"rn":      {"mesh_headings"},
	"kf":      {"mesh_headings"},
	"sb":      {"mesh_headings"},
	"mh":      {"mesh_headings"},
	"pt":      {"pub_type"},
	"au":      {"author"},
	"default": {"abstract"},
}

var adjMatchRegexp, _ = regexp.Compile("^adj[0-9]*$")

// MedlineTransformer is an implementation of a QueryTransformer in the parser package.
type MedlineTransformer struct{}

// TransformFields maps a string of fields into a slice of mapped fields.
func (p MedlineTransformer) TransformFields(fields string, mapping map[string][]string) []string {
	parts := strings.Split(fields, ",")
	mappedFields := []string{}
	for _, field := range parts {
		mappedFields = append(mappedFields, mapping[field]...)
	}
	return mappedFields
}

// TransformNested implements the transformation of a nested query.
func (p MedlineTransformer) TransformNested(query string, mapping map[string][]string) ir.BooleanQuery {
	var fieldsString string
	for i := len(query) - 1; i > 0; i-- {
		if query[i] == ')' {
			break
		}
		fieldsString += string(query[i])
	}
	fieldsString = ReversePreservingCombiningCharacters(fieldsString)
	query = strings.Replace(query, fieldsString, "", 1)

	fields := []string{}
	fieldsString = strings.Replace(fieldsString, ".", "", -1)
	for _, field := range strings.Split(fieldsString, ",") {
		fields = append(fields, mapping[field]...)
	}

	// Add a default field to the keyword if none have been defined
	if len(fields) == 0 {
		fields = mapping["default"]
	}

	return p.ParseInfixKeywords(query, fields, mapping)
}

// TransformSingle implements the transformation of a single, stand-alone query. This is called from TransformNested
// to transform the inner queries.
func (p MedlineTransformer) TransformSingle(query string, mapping map[string][]string) ir.Keyword {
	var queryString string
	var fields []string
	exploded := false

	if query[len(query)-1] == '/' {
		// Check to see if we are looking at a mesh heading string.
		expParts := strings.Split(query, " ")
		if expParts[0] == "exp" {
			queryString = strings.Join(expParts[1:], " ")
			exploded = true
		} else {
			queryString = query
		}
		queryString = strings.Replace(queryString, "/", "", -1)
		fields = mapping["sh"]
	} else {
		// Otherwise try to parse a regular looking query.
		parts := strings.Split(query, ".")
		if len(parts) == 3 {
			queryString = parts[0]
			fields = p.TransformFields(parts[1], mapping)
		} else {
			queryString = query
		}
	}

	// medline uses $ to represent the stem of a word. Instead let's just replace it by the wildcard operator.
	// TODO is there anything in Elasticsearch to do this?
	queryString = strings.Replace(queryString, "$", "*", -1)

	queryString = strings.TrimSpace(queryString)

	return ir.Keyword{
		QueryString: queryString,
		Fields:      fields,
		Exploded:    exploded,
		Truncated:   false,
	}
}

// transformPrefixGroupToQueryGroup transforms a prefix syntax tree into a query group. The new QueryGroup is built by
// recursively navigating the syntax tree.
func (p MedlineTransformer) TransformPrefixGroupToQueryGroup(prefix []string, queryGroup ir.BooleanQuery, fields []string, mapping map[string][]string) ([]string, ir.BooleanQuery) {
	//log.Println(queryGroup)
	if len(prefix) == 0 {
		return prefix, queryGroup
	}

	token := prefix[0]
	if p.IsOperator(token) {
		queryGroup.Operator = token
	} else if token == "(" {
		var subGroup ir.BooleanQuery
		prefix, subGroup = p.TransformPrefixGroupToQueryGroup(prefix[1:], ir.BooleanQuery{}, fields, mapping)
		queryGroup.Children = append(queryGroup.Children, subGroup)
	} else if token == ")" {
		return prefix, queryGroup
	} else {
		if len(token) > 0 {
			k := p.TransformSingle(token, mapping)
			// Add a default field to the keyword if none have been defined
			if len(k.Fields) == 0 && len(fields) > 0 {
				k.Fields = fields
			} else if len(k.Fields) == 0 && len(fields) == 0 {
				log.Printf("no inner or outer fields are defined for nested query `%v`, using default (%v)", token, mapping["default"])
				k.Fields = mapping["default"]
			}
			queryGroup.Keywords = append(queryGroup.Keywords, k)
		}
	}
	return p.TransformPrefixGroupToQueryGroup(prefix[1:], queryGroup, fields, mapping)
}

// ConvertInfixToPrefix translates an infix grouping expression into a prefix expression. The way this is done is the
// Shunting-yard algorithm (https://en.wikipedia.org/wiki/Shunting-yard_algorithm).
func (p MedlineTransformer) ConvertInfixToPrefix(infix []string) []string {
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

// ParseInfixKeywords parses an infix expression containing keywords separated by operators into an infix expression,
// and then into the immediate representation.
func (p MedlineTransformer) ParseInfixKeywords(line string, fields []string, mapping map[string][]string) ir.BooleanQuery {
	line += "\n"

	stack := []string{}

	keyword := ""
	currentToken := ""
	previousToken := ""

	endTokens := ""

	depth := 0

	for _, char := range line {
		if unicode.IsSpace(char) {
			t := strings.ToLower(currentToken)
			if p.IsOperator(t) {
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
				stack = append(stack, strings.TrimSpace(keyword+" "+previousToken+" "+currentToken))
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
	prefix := p.ConvertInfixToPrefix(stack)
	if prefix[0] == "(" && prefix[len(prefix)-1] == ")" {
		prefix = prefix[1: len(prefix)-1]
	}
	_, queryGroup := p.TransformPrefixGroupToQueryGroup(prefix, ir.BooleanQuery{}, fields, mapping)
	return queryGroup
}

// reversePreservingCombiningCharacters interprets its argument as UTF-8
// and ignores bytes that do not form valid UTF-8.  return value is UTF-8.
// https://rosettacode.org/wiki/Reverse_a_string#Go
func ReversePreservingCombiningCharacters(s string) string {
	if s == "" {
		return ""
	}
	p := []rune(s)
	r := make([]rune, len(p))
	start := len(r)
	for i := 0; i < len(p); {
		// quietly skip invalid UTF-8
		if p[i] == utf8.RuneError {
			i++
			continue
		}
		j := i + 1
		for j < len(p) && (unicode.Is(unicode.Mn, p[j]) ||
			unicode.Is(unicode.Me, p[j]) || unicode.Is(unicode.Mc, p[j])) {
			j++
		}
		for k := j - 1; k >= i; k-- {
			start--
			r[start] = p[k]
		}
		i = j
	}
	return string(r[start:])
}

// IsOperator tests to see if a string is a valid PubMed/Medline operator.
func (p MedlineTransformer) IsOperator(s string) bool {
	return s == "or" ||
		s == "and" ||
		s == "not" ||
		adjMatchRegexp.MatchString(s)
}

func NewMedlineParser() QueryParser {
	return QueryParser{FieldMapping: MedlineFieldMapping, Parser: MedlineTransformer{}}
}
