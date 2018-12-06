package backend

import (
	"bytes"
	"fmt"
	"github.com/hscells/cqr"
	"github.com/hscells/transmute/fields"
	"github.com/hscells/transmute/ir"
	"sort"
	"strings"
)

type PubmedBackend struct {
	ReplaceAdj bool
}

type PubmedQuery struct {
	repr string
}

func (m PubmedQuery) Representation() (interface{}, error) {
	return m.repr, nil
}

func (m PubmedQuery) String() (string, error) {
	return m.repr, nil
}

func (m PubmedQuery) StringPretty() (string, error) {
	return m.repr, nil
}

func compilePubmed(q ir.BooleanQuery, level int, replaceAdj bool) (l int, query PubmedQuery) {
	if q.Keywords == nil && len(q.Operator) == 0 {
		repr := ""
		for _, child := range q.Children {
			var comp PubmedQuery
			level, comp = compilePubmed(child, level, replaceAdj)
			repr += comp.repr
		}
		return level, PubmedQuery{repr: repr}
	}

	children := make([]string, len(q.Children))
	for i, child := range q.Children {
		l, comp := compilePubmed(child, level, replaceAdj)
		level = l
		children[i] = comp.repr
	}
	keywords := make([]string, len(q.Keywords))
	for i, keyword := range q.Keywords {
		var mf string
		qs := keyword.QueryString
		buff := new(bytes.Buffer)

		// PubMed supports only end-truncation. There is no single character symbol.
		// https://www.nlm.nih.gov/bsd/disted/pubmedtutorial/020_460.html
		for i, char := range qs {
			if i > 0 && (char == '?' || char == '$' || char == '*') {
				buff.WriteRune('*')
				if qs[0] == '"' {
					buff.WriteRune('"')
				}
				qs = buff.String()
				break
			} else if i == 0 && (char == '?' || char == '$' || char == '*') {
				continue
			}
			buff.WriteRune(char)
		}

		if len(keyword.Fields) == 1 {
			if keyword.Fields[0] == fields.MeshHeadings {
				mf = "Mesh Terms"
			} else if keyword.Fields[0] == fields.FloatingMeshHeadings {
				mf = "MeSH Subheading"
			} else if keyword.Fields[0] == fields.MajorFocusMeshHeading {
				mf = "MeSH Major Topic"
			}
			if len(mf) > 0 && !keyword.Exploded {
				mf += ":noexp"
			}
		}

		if len(mf) == 0 {
			mapping1 := map[string][]string{
				"Affiliation":                     {fields.Affiliation},
				"All Fields":                      {fields.AllFields},
				"Author":                          {fields.Author},
				"Authors":                         {fields.Authors},
				"Author - Corporate":              {fields.AuthorCorporate},
				"Author - First":                  {fields.AuthorFirst},
				"Author - Full":                   {fields.AuthorFull},
				"Author - Identifier":             {fields.AuthorIdentifier},
				"Author - Last":                   {fields.AuthorLast},
				"Book":                            {fields.Book},
				"Date - Completion":               {fields.DateCompletion},
				"Conflict Of Interest Statements": {fields.ConflictOfInterestStatements},
				"Date - Create":                   {fields.DateCreate},
				"Date - Entrez":                   {fields.DateEntrez},
				"Date - MeSH":                     {fields.DateMeSH},
				"Date - Modification":             {fields.DateModification},
				"Date - Publication":              {fields.DatePublication},
				"EC/RN Number":                    {fields.ECRNNumber},
				"Editor":                          {fields.Editor},
				"Filter":                          {fields.Filter},
				"Grant Number":                    {fields.GrantNumber},
				"ISBN":                            {fields.ISBN},
				"Investigator":                    {fields.Investigator},
				"Investigator - Full":             {fields.InvestigatorFull},
				"Issue":                           {fields.Issue},
				"Journal":                         {fields.Journal},
				"Language":                        {fields.Language},
				"Location ID":                     {fields.LocationID},
				"MeSH Major Topic":                {fields.MeSHMajorTopic},
				"MeSH Subheading":                 {fields.MeSHSubheading},
				"MeSH Terms":                      {fields.MeSHTerms},
				"Other Term":                      {fields.OtherTerm},
				"Pagination":                      {fields.Pagination},
				"Pharmacological Action":          {fields.PharmacologicalAction},
				"Publication Type":                {fields.PublicationType},
				"Publisher":                       {fields.Publisher},
				"Secondary Source ID":             {fields.SecondarySourceID},
				"Subject Personal Name":           {fields.SubjectPersonalName},
				"Supplementary Concept":           {fields.SupplementaryConcept},
				"Floating MeshHeadings":           {fields.FloatingMeshHeadings},
				"Text Word":                       {fields.TextWord},
				"Title":                           {fields.Title},
				"Title/Abstract":                  {fields.TitleAbstract},
				"Transliterated Title":            {fields.TransliteratedTitle},
				"Volume":                          {fields.Volume},
				"MeSH Headings":                   {fields.MeshHeadings},
				"Major Focus MeSH Heading":        {fields.MajorFocusMeshHeading},
				"Publication Date":                {fields.PublicationDate},
				"Publication Status":              {fields.PublicationStatus},
			}
			mapping2 := map[string][]string{
				"Title/Abstract": {fields.Abstract, fields.Title},
				"Text Word":      {fields.Abstract},
			}
			sort.Strings(keyword.Fields)
			for f, mappingFields := range mapping1 {
				if len(mappingFields) != len(keyword.Fields) {
					continue
				}
				match := true
				for i, field := range keyword.Fields {
					if field != mappingFields[i] {
						match = false
					}
				}
				if match {
					mf = f
					break
				}
			}
			for f, mappingFields := range mapping2 {
				if len(mappingFields) != len(keyword.Fields) {
					continue
				}
				match := true
				for i, field := range keyword.Fields {
					if field != mappingFields[i] {
						match = false
					}
				}
				if match {
					mf = f
					break
				}
			}
			// This should be a sensible enough default.
			if len(mf) == 0 {
				mf = "All Fields"
			}
		}
		qs = fmt.Sprintf("%v[%v]", qs, mf)
		keywords[i] = qs
		level += 1
	}

	keywords = append(keywords, children...)

	if strings.Contains(strings.ToLower(q.Operator), "adj") {
		q.Operator = cqr.AND
	}

	repr := fmt.Sprintf("(%v)", strings.Join(keywords, strings.ToUpper(fmt.Sprintf(" %v ", q.Operator))))
	level += 1
	return level, PubmedQuery{repr: repr}
}

func (b PubmedBackend) Compile(ir ir.BooleanQuery) (BooleanQuery, error) {
	_, q := compilePubmed(ir, 1, b.ReplaceAdj)
	return q, nil
}

func NewPubmedBackend() PubmedBackend {
	return PubmedBackend{}
}
