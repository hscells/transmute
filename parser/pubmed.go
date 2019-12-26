package parser

import (
	"github.com/hscells/transmute/fields"
	"github.com/hscells/transmute/ir"
	"log"
	"strings"
	"unicode"
)

type PubMedTransformer struct{}

var PubMedFieldMapping = map[string][]string{
	"Mesh":                              {fields.MeshHeadings},
	"mesh":                              {fields.MeshHeadings},
	"MeSH":                              {fields.MeshHeadings},
	"MESH":                              {fields.MeshHeadings},
	"Mesh Terms":                        {fields.MeshHeadings},
	"mesh terms":                        {fields.MeshHeadings},
	"MAJR":                              {fields.MajorFocusMeshHeading},
	"mesh major topic":                  {fields.MajorFocusMeshHeading},
	"Subheading":                        {fields.FloatingMeshHeadings},
	"subheading":                        {fields.FloatingMeshHeadings},
	"Abstract":                          {fields.Abstract},
	"Publication":                       {fields.PublicationType},
	"publication type":                  {fields.PublicationType},
	"journal":                           {fields.Journal},
	"Date - Entrez : 3000":              {fields.PublicationDate},
	"Publication Date":                  {fields.PublicationDate},
	"Affiliation":                       {fields.Affiliation},
	"All Fields":                        {fields.AllFields},
	"Author":                            {fields.Author},
	"Author - Corporate":                {fields.AuthorCorporate},
	"Author - First":                    {fields.AuthorFirst},
	"Author - Full":                     {fields.AuthorFull},
	"Author - Identifier":               {fields.AuthorIdentifier},
	"Author - Last":                     {fields.AuthorLast},
	"Book":                              {fields.Book},
	"Date - Completion":                 {fields.DateCompletion},
	"Conflict of Interest Statements":   {fields.ConflictOfInterestStatements},
	"Date - Create":                     {fields.DateCreate},
	"Date - Entrez":                     {fields.DateEntrez},
	"Date - MeSH":                       {fields.DateMeSH},
	"Date - Modification":               {fields.DateModification},
	"Date - Publication":                {fields.DatePublication},
	"EC/RN Number":                      {fields.ECRNNumber},
	"Editor":                            {fields.Editor},
	"Filter":                            {fields.Filter},
	"Grant Number":                      {fields.GrantNumber},
	"ISBN":                              {fields.ISBN},
	"Investigator":                      {fields.Investigator},
	"Investigator - Full":               {fields.InvestigatorFull},
	"Issue":                             {fields.Issue},
	"Journal":                           {fields.Journal},
	"Language":                          {fields.Language},
	"Location ID":                       {fields.LocationID},
	"MeSH Major Topic":                  {fields.MeSHMajorTopic},
	"MeSH Subheading":                   {fields.MeSHSubheading},
	"MeSH Terms":                        {fields.MeSHTerms},
	"Other Term":                        {fields.OtherTerm},
	"Pagination":                        {fields.Pagination},
	"Pharmacological Action":            {fields.PharmacologicalAction},
	"Publication Type":                  {fields.PublicationType},
	"Publisher":                         {fields.Publisher},
	"Secondary Source ID":               {fields.SecondarySourceID},
	"Subject - Personal Name":           {fields.SubjectPersonalName},
	"Supplementary Concept":             {fields.SupplementaryConcept},
	"Text Word":                         {fields.TextWord},
	"Title":                             {fields.Title},
	"Title/Abstract":                    {fields.TitleAbstract},
	"Transliterated Title":              {fields.TransliteratedTitle},
	"Volume":                            {fields.Volume},
	"mh":                                {fields.MeshHeadings},
	"sh":                                {fields.FloatingMeshHeadings},
	"tw":                                {fields.TextWord},
	"ti":                                {fields.Title},
	"pt":                                {fields.PublicationType},
	"sb":                                {fields.PublicationStatus},
	"tiab":                              {fields.TitleAbstract},
	"text":                              {fields.TitleAbstract},
	fields.Affiliation:                  {fields.Affiliation},
	fields.AllFields:                    {fields.AllFields},
	fields.Author:                       {fields.Author},
	fields.Authors:                      {fields.Authors},
	fields.AuthorCorporate:              {fields.AuthorCorporate},
	fields.AuthorFirst:                  {fields.AuthorFirst},
	fields.AuthorFull:                   {fields.AuthorFull},
	fields.AuthorIdentifier:             {fields.AuthorIdentifier},
	fields.AuthorLast:                   {fields.AuthorLast},
	fields.Book:                         {fields.Book},
	fields.DateCompletion:               {fields.DateCompletion},
	fields.ConflictOfInterestStatements: {fields.ConflictOfInterestStatements},
	fields.DateCreate:                   {fields.DateCreate},
	fields.DateEntrez:                   {fields.DateEntrez},
	fields.DateMeSH:                     {fields.DateMeSH},
	fields.DateModification:             {fields.DateModification},
	fields.DatePublication:              {fields.DatePublication},
	fields.ECRNNumber:                   {fields.ECRNNumber},
	fields.Editor:                       {fields.Editor},
	fields.Filter:                       {fields.Filter},
	fields.GrantNumber:                  {fields.GrantNumber},
	fields.ISBN:                         {fields.ISBN},
	fields.Investigator:                 {fields.Investigator},
	fields.InvestigatorFull:             {fields.InvestigatorFull},
	fields.Issue:                        {fields.Issue},
	fields.Journal:                      {fields.Journal},
	fields.Language:                     {fields.Language},
	fields.LocationID:                   {fields.LocationID},
	fields.MeSHMajorTopic:               {fields.MeSHMajorTopic},
	fields.MeSHSubheading:               {fields.MeSHSubheading},
	fields.MeSHTerms:                    {fields.MeSHTerms},
	fields.OtherTerm:                    {fields.OtherTerm},
	fields.Pagination:                   {fields.Pagination},
	fields.PharmacologicalAction:        {fields.PharmacologicalAction},
	fields.PublicationType:              {fields.PublicationType},
	fields.Publisher:                    {fields.Publisher},
	fields.SecondarySourceID:            {fields.SecondarySourceID},
	fields.SubjectPersonalName:          {fields.SubjectPersonalName},
	fields.SupplementaryConcept:         {fields.SupplementaryConcept},
	fields.FloatingMeshHeadings:         {fields.FloatingMeshHeadings},
	fields.TextWord:                     {fields.TextWord},
	fields.Title:                        {fields.Title},
	fields.TitleAbstract:                {fields.TitleAbstract},
	fields.TransliteratedTitle:          {fields.TransliteratedTitle},
	fields.Volume:                       {fields.Volume},
	fields.MeshHeadings:                 {fields.MeshHeadings},
	fields.MajorFocusMeshHeading:        {fields.MajorFocusMeshHeading},
	fields.PublicationDate:              {fields.PublicationDate},
	fields.PublicationStatus:            {fields.PublicationStatus},
	fields.PMID:                         {fields.PMID},
	"default":                           {fields.AllFields},
}

func (t PubMedTransformer) TransformSingle(query string, mapping map[string][]string) ir.Keyword {
	var queryString string
	var queryFields []string
	exploded := true

	if strings.ContainsRune(query, '[') {
		// This query string most likely has a field.
		parts := strings.Split(query, "[")
		queryString = parts[0]
		// This might be a field, but needs some processing.
		possibleField := strings.Replace(parts[1], "]", "", -1)

		// Set the exploded option on the keyword.
		if strings.Contains(strings.ToLower(possibleField), "mesh") || strings.Contains(strings.ToLower(possibleField), "heading") {
			exploded = true
		}

		// PubMed fields have this weird thing where they specify the mesh explosion in the field.
		// This is handled in this step.
		if strings.Contains(strings.ToLower(possibleField), ":noexp") {
			exploded = false
			possibleField = strings.Replace(strings.ToLower(possibleField), ":noexp", "", -1)
		}

		// If we are unable to map the field then we can explode.
		if field, ok := mapping[possibleField]; ok {
			queryFields = field
		} else {
			log.Printf("the field `%v` does not have a mapping defined\n", possibleField)
			queryFields = mapping["default"]
		}
	} else {
		queryString = query
	}

	// Add a default field to the keyword if none have been defined.
	if len(queryFields) == 0 {
		log.Printf("using default field (%v) since %v has no queryFields\n", mapping["default"], query)
		queryFields = mapping["default"]
	}

	// PubMed uses $ to represent the stem of a word. Instead let's just replace it by the wildcard operator.
	truncated := false
	if strings.ContainsAny(queryString, "*$?~") {
		truncated = true
	}
	queryString = strings.Replace(queryString, "$", "*", -1)
	queryString = strings.Replace(queryString, "?", "*", -1)
	queryString = strings.Replace(queryString, "~", "*", -1)
	//queryString = strings.Replace(queryString, "*", " ", -1)

	queryString = strings.TrimSpace(queryString)

	return ir.Keyword{
		QueryString: queryString,
		Fields:      queryFields,
		Exploded:    exploded,
		Truncated:   truncated,
	}
}

func (t PubMedTransformer) TransformNested(query string, mapping map[string][]string) ir.BooleanQuery {
	query = ReversePreservingCombiningCharacters(reverse(query))
	return t.ParseInfixKeywords(query, mapping)
}

func (t PubMedTransformer) RemoveParenthesis(expr []string) []string {
	r := make([]string, len(expr))
	s := make([]string, len(expr))

	copy(r, expr)
	copy(s, expr)

	var st []int
	i := 0
	for i < len(s) {
		if s[i] == "(" {
			if s[i+1] == "(" {
				st = append(st, -i)
			} else {
				st = append(st, i)
			}
			i++
		} else if s[i] != ")" && s[i] != "(" {
			i++
		} else if s[i] == ")" {
			top := st[len(st)-1]
			if s[i-1] == ")" && top < 0 {
				r[-top] = "$"
				r[i] = "$"
				st = st[:len(st)-1]
			} else if s[i-1] == ")" && top > 0 {
				//panic("invalid query")
			} else if s[i-1] != ")" && top > 0 {
				st = st[:len(st)-1]
			}
			i++
		}
	}

	var result []string
	for i := 0; i < len(r); i++ {
		if r[i] == "$" {
			continue
		}
		result = append(result, r[i])
	}

	return result
}

// ParseInfixKeywords parses an infix expression containing keywords separated by operators into an infix expression,
// and then into the immediate representation.
func (t PubMedTransformer) ParseInfixKeywords(line string, mapping map[string][]string) ir.BooleanQuery {
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
				stack = append(stack, strings.TrimSpace(keyword+" "+previousToken+" "+currentToken))
				keyword = ""
				currentToken = ""
				previousToken = ""
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
		prefix = prefix[1 : len(prefix)-1]
	}

	prefix = append([]string{"("}, prefix...)
	prefix = append(prefix, ")")

	//fmt.Println(prefix)
	prefix = t.RemoveParenthesis(prefix)

	// Remove redundancy.
	var p []string
	var prev string
	for _, token := range prefix {
		if prev == token && token != ")" && token != "(" {
			continue
		}
		prev = token
		p = append(p, token)
	}

	var l []string
	var tmp []string
	inside := false
	for _, token := range p {
		if prev == "(" && token == "(" && !inside {
			inside = true
			prev = token
			continue
		}
		if inside {
			if prev == ")" && token == ")" {
				l = append(l, tmp...)
				tmp = []string{}
				prev = token
				inside = false
				continue
			}
			tmp = append(tmp, token)
			prev = token
		} else {
			l = append(l, token)
			prev = token
		}
	}

	_, queryGroup := t.TransformPrefixGroupToQueryGroup(l, ir.BooleanQuery{}, mapping)
	return queryGroup
}

// ConvertInfixToPrefix translates an infix grouping expression into a prefix expression. The way this is done is the
// Shunting-yard algorithm (https://en.wikipedia.org/wiki/Shunting-yard_algorithm).
func (t PubMedTransformer) ConvertInfixToPrefix(infix []string) []string {
	// The stack contains some intermediate values
	var stack []string
	// The result contains the actual expression
	var result []string

	precedence := map[string]int{
		"and": 0,
		"or":  1,
		"not": 2,
	}

	// The algorithm is slightly modified to also store the brackets in the result
	for i := len(infix) - 1; i >= 0; i-- {
		token := infix[i]
		if len(token) == 0 {
			continue
		}
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
	if len(prefix) <= 1 {
		return prefix, queryGroup
	}

	token := prefix[0]
	if t.IsOperator(token) {
		queryGroup.Operator = token
	} else if token == "(" {
		var subGroup ir.BooleanQuery
		prefix, subGroup = t.TransformPrefixGroupToQueryGroup(prefix[1:], ir.BooleanQuery{}, mapping)
		if len(prefix) == 0 {
			return prefix, queryGroup
		}
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
