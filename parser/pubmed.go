package parser

import (
	"github.com/hscells/transmute/ir"
	"strings"
)

var PubMedFieldMapping = map[string][]string{
	"mp": {"title", "abstract", "mesh_headings"},
	"af": {"title", "abstract", "mesh_headings"},
	"tw": {"title", "abstract"},
	"nm": {"abstract", "mesh_headings"},
	"ab": {"abstract"},
	"ti": {"title"},
	"ot": {"title"},
	"sh": {"mesh_headings"},
	"px": {"mesh_headings"},
	"rs": {"mesh_headings"},
	"fs": {"mesh_headings"},
	"rn": {"mesh_headings"},
	"kf": {"mesh_headings"},
	"sb": {"mesh_headings"},
	"mh": {"mesh_headings"},
	"pt": {"pubtype"},
}

type PubMedParser struct{}

func TransformFields(fields string) []string {
	parts := strings.Split(fields, ",")
	mappedFields := []string{}
	for _, field := range parts {
		mappedFields = append(mappedFields, PubMedFieldMapping[field]...)
	}
	return mappedFields
}

func (p PubMedParser) Transform(query string) ir.Keyword {
	parts := strings.Split(query, ".")
	var queryString string
	var fields []string
	if len(parts) == 3 {
		queryString = parts[0]
		fields = TransformFields(parts[1])
	} else {
		queryString = query
	}
	return ir.Keyword{
		QueryString: queryString,
		Fields:      fields,
		Exploded:    false,
		Truncated:   false,
	}
}

func NewPubMedParser() QueryParser {
	return QueryParser{FieldMapping: PubMedFieldMapping, Parser: PubMedParser{}}
}
