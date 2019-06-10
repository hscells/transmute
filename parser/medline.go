package parser

import (
	"fmt"
	"github.com/hscells/transmute/fields"
	"github.com/hscells/transmute/ir"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

var MedlineFieldMapping = map[string][]string{
	"ab":      {fields.Abstract},
	"af":      {fields.AllFields},
	"ai":      {fields.AuthorFull},
	"as":      {fields.PublicationDate},
	"au":      {fields.Authors},
	"ax":      {fields.AuthorLast},
	"ba":      {fields.Authors},
	"bd":      {fields.PublicationDate},
	"be":      {fields.Editor},
	"bf":      {fields.Authors},
	"bk":      {fields.AllFields},
	"em":      {fields.PublicationDate},
	"ed":      {fields.PublicationDate},
	"fa":      {fields.AuthorFull},
	"fe":      {fields.Editor},
	"fs":      {fields.FloatingMeshHeadings},
	"fx":      {fields.FloatingMeshHeadings},
	"kf":      {fields.AllFields},
	"ot":      {fields.Title},
	"mp":      {fields.AllFields},
	"mh":      {fields.MeshHeadings},
	"nm":      {fields.AllFields},
	"px":      {fields.MeshHeadings},
	"pt":      {fields.PublicationType},
	"rs":      {fields.AllFields},
	"rn":      {fields.AllFields},
	"sb":      {fields.PublicationType},
	"sh":      {fields.MeSHSubheading},
	"tw":      {fields.TextWord},
	"ti":      {fields.Title},
	"ja":      {fields.Journal},
	"jn":      {fields.Journal},
	"jw":      {fields.Journal},
	"ti,ab":   {fields.TitleAbstract},
	"default": {fields.AllFields},
}

var adjMatchRegexp, _ = regexp.Compile("^adj[0-9]*$")
var medlineFieldRegexp, _ = regexp.Compile(".[a-z]{2}.")

// MedlineTransformer is an implementation of a QueryTransformer in the parser package.
type MedlineTransformer struct{}

// TransformFields maps a string of fields into a slice of mapped fields.
func (p MedlineTransformer) TransformFields(fields string, mapping map[string][]string) []string {
	parts := strings.Split(fields, ",")
	var mappedFields []string
	for _, field := range parts {
		mappedFields = append(mappedFields, mapping[field]...)
	}
	sort.Strings(mappedFields)
	return mappedFields
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
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
	fieldsString = reverse(fieldsString)

	fieldsString = ReversePreservingCombiningCharacters(fieldsString)
	query = strings.Replace(query, fieldsString, "", 1)

	var queryFields []string
	fieldsString = strings.Replace(fieldsString, ".", "", -1)
	for _, field := range strings.Split(fieldsString, ",") {
		queryFields = append(queryFields, mapping[field]...)
	}

	// Add a default field to the keyword if none have been defined
	if len(queryFields) == 0 {
		queryFields = mapping["default"]
	}

	return p.ParseInfixKeywords(query, queryFields, mapping)
}

// TransformSingle implements the transformation of a single, stand-alone query. This is called from TransformNested
// to transform the inner queries.
func (p MedlineTransformer) TransformSingle(query string, mapping map[string][]string) ir.Keyword {
	var queryString string
	var queryFields []string
	exploded := false

	// Trim the query string to prevent whitespace such as newlines interfering with string processing.
	query = strings.TrimSpace(query)

	if len(query) > 0 && query[len(query)-1] == '/' {
		// Check to see if we are looking at a mesh heading string.
		expParts := strings.Split(query, " ")
		if expParts[0] == "exp" {
			queryString = strings.Join(expParts[1:], " ")
			exploded = true
		} else {
			queryString = query
		}
		queryString = strings.Replace(queryString, "/", "", -1)
		queryFields = mapping["sh"]
	} else {
		// Otherwise try to parse a regular looking query.
		parts := strings.Split(query, ".")
		if len(parts) > 1 {
			queryString = strings.Join(parts[0:len(parts)-2], ".")
			queryFields = p.TransformFields(parts[len(parts)-2], mapping)
		} else {
			queryString = query
		}
	}

	truncated := false
	if strings.ContainsAny(queryString, "*$?~") {
		truncated = true
	}

	queryString = strings.Replace(queryString, "$", "*", -1)
	queryString = strings.Replace(queryString, "~", "*", -1)

	queryString = strings.TrimSpace(queryString)

	return ir.Keyword{
		QueryString: queryString,
		Fields:      queryFields,
		Exploded:    exploded,
		Truncated:   truncated,
	}
}

// transformPrefixGroupToQueryGroup transforms a prefix syntax tree into a query group. The new QueryGroup is built by
// recursively navigating the syntax tree.
func (p MedlineTransformer) TransformPrefixGroupToQueryGroup(prefix []string, queryGroup ir.BooleanQuery, fields []string, mapping map[string][]string) ([]string, ir.BooleanQuery) {
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
		// At this point, the next item in the prefix slice can be the fields for the inner query terms.
		// Ths needs to be handled!!
		// Process the default fields.
		foundFields := mapping["default"]
		if len(prefix) > 1 {
			// When we have a prefix that matches a field for the previous inner group of queries.
			if medlineFieldRegexp.MatchString(prefix[1]) {
				fieldString := prefix[1][1:3]

				// We can try to map them.
				if strings.Contains(fieldString, ",") {
					foundFields = p.TransformFields(fieldString, mapping)
				} else {
					if f, ok := mapping[fieldString]; ok {
						foundFields = f
					}
				}

				prefix = prefix[1:]
			}
		}
		for i, kw := range queryGroup.Keywords {
			if kw.Fields == nil || len(kw.Fields) == 0 {
				queryGroup.Keywords[i].Fields = foundFields
			}
		}

		return prefix, queryGroup
	} else {
		if len(token) > 0 {
			k := p.TransformSingle(token, mapping)
			if len(k.Fields) == 0 && prefix[len(prefix)-1] != ")" {
				token = fmt.Sprintf("%s%s", token, prefix[len(prefix)-1])
				k = p.TransformSingle(token, mapping)
			}
			// Add a default field to the keyword if none have been defined
			//if len(k.Fields) == 0 && len(fields) > 0 {
			//	k.Fields = fields
			//} else if len(k.Fields) == 0 && len(fields) == 0 {
			//	log.Printf("no inner or outer fields are defined for nested query `%v`, using default (%v)", token, mapping["default"])
			//	k.Fields = mapping["default"]
			//}
			if len(k.QueryString) > 0 {
				queryGroup.Keywords = append(queryGroup.Keywords, k)
			}
		}
	}
	pf := prefix
	if len(prefix) > 1 {
		pf = prefix[1:]
	} else {
		pf = []string{}
	}
	return p.TransformPrefixGroupToQueryGroup(pf, queryGroup, fields, mapping)
}

// ConvertInfixToPrefix translates an infix grouping expression into a prefix expression. The way this is done is the
// Shunting-yard algorithm (https://en.wikipedia.org/wiki/Shunting-yard_algorithm).
func (p MedlineTransformer) ConvertInfixToPrefix(infix []string) []string {
	// The stack contains some intermediate values
	var stack []string
	// The result contains the actual expression
	var result []string

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

	var stack []string

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
			currentToken += `"`
			continue
		} else if char == '"' && insideQuote {
			insideQuote = false
			currentToken += `"`
			continue
		} else if insideQuote {
			currentToken += string(char)
			continue
		}

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
	if len(prefix) > 0 {
		if prefix[0] == "(" && prefix[len(prefix)-1] == ")" {
			prefix = prefix[1 : len(prefix)-1]
		}
	}
	_, queryGroup := p.TransformPrefixGroupToQueryGroup(prefix, ir.BooleanQuery{}, fields, mapping)
	return queryGroup
}

// ReversePreservingCombiningCharacters interprets its argument as UTF-8
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
