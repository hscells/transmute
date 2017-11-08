package parser

import (
	"bytes"
	"encoding/json"
	"github.com/hscells/transmute/ir"
	"log"
)

// CQRTransformer is an implementation of a query transformer for CQR queries.
type CQRTransformer struct{}

// TransformSingle is unused for this parser.
func (c CQRTransformer) TransformSingle(query string, mapping map[string][]string) ir.Keyword {
	return ir.Keyword{}
}

// transformSingle maps CQR keywords to ir keywords.
func transformSingle(rep map[string]interface{}) ir.Keyword {
	fields := []string{}
	for _, field := range rep["fields"].([]interface{}) {
		fields = append(fields, field.(string))
	}
	var exploded, truncated bool
	if v, ok := rep["options"].(map[string]interface{})["exploded"]; ok {
		exploded = v.(bool)
	}
	if v, ok := rep["options"].(map[string]interface{})["truncated"]; ok {
		truncated = v.(bool)
	}

	return ir.Keyword{
		QueryString: rep["query"].(string),
		Fields:      fields,
		Exploded:    exploded,
		Truncated:   truncated,
	}
}

// transformNested transforms the CQR nested queries.
func transformNested(rep map[string]interface{}) ir.BooleanQuery {
	q := ir.BooleanQuery{Children: []ir.BooleanQuery{}, Keywords: []ir.Keyword{}}
	if rep["children"] != nil {
		q.Operator = rep["operator"].(string)
		for _, child := range rep["children"].([]interface{}) {
			cq := child.(map[string]interface{})
			if _, ok := cq["operator"]; !ok {
				q.Keywords = append(q.Keywords, transformSingle(cq))
			} else {
				q.Children = append(q.Children, transformNested(cq))
			}
		}
	}

	return q
}

// TransformNested takes a JSON string a parses a CQR object into the ir.
func (c CQRTransformer) TransformNested(query string, mapping map[string][]string) ir.BooleanQuery {
	var queryRep map[string]interface{}
	err := json.Unmarshal(bytes.NewBufferString(query).Bytes(), &queryRep)
	if err != nil {
		log.Fatalln(err)
	}

	return transformNested(queryRep)
}

// NewCQRParser creates a new parser for CQR queries. This parser makes a lot of assumptions as it assumes the
// structure of this query is perfect.
func NewCQRParser() QueryParser {
	return QueryParser{Parser: CQRTransformer{}}
}
