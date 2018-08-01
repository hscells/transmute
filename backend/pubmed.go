package backend

import (
	"github.com/hscells/transmute/ir"
	"fmt"
	"strings"
	"github.com/hscells/cqr"
	"bytes"
	"sort"
	"github.com/hscells/transmute/fields"
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
				mf = "Subheading"
			} else if keyword.Fields[0] == fields.MajorFocusMeshHeading {
				mf = "MeSH Major Topic"
			}
			if len(mf) > 0 && !keyword.Exploded {
				mf += ":noexp"
			}
		}

		if len(mf) == 0 {
			mapping := map[string][]string{
				"Mesh Terms":       {fields.MeshHeadings},
				"Subheading":       {fields.FloatingMeshHeadings},
				"MeSH Major Topic": {fields.MajorFocusMeshHeading},
				"Title/Abstract":   {fields.Abstract, fields.Title},
				"Title":            {fields.Title},
				"Text Word":        {fields.Abstract},
				"Publication Type": {fields.PublicationType},
				"Publication Date": {fields.PublicationDate},
				"Journal":          {fields.Journal},
			}
			sort.Strings(keyword.Fields)
			for f, mappingFields := range mapping {
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
