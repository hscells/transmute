package backend

import (
	"github.com/hscells/transmute/ir"
	"fmt"
	"strings"
)

type TerrierQuery struct {
	repr string
}

type TerrierBackend struct{}

//var (
//regexSpace, _ = regexp.Compile(" +")
//)

// String returns a JSON-encoded representation of the cqr.
func (q TerrierQuery) String() string {
	return q.repr
}

// StringPretty returns a pretty-printed JSON-encoded representation of the cqr.
func (q TerrierQuery) StringPretty() string {
	return q.String()
}

func (q TerrierQuery) Representation() interface{} {
	return q.String()
}

func (t TerrierBackend) Compile(q ir.BooleanQuery) BooleanQuery {
	tq := TerrierQuery{}

	// Process the keywords.
	if q.Operator == "and" {
		tq.repr = "("
		keywords := []string{}
		for _, keyword := range q.Keywords {
			for _, field := range keyword.Fields {
				keywords = append(keywords, fmt.Sprintf("+%s:%s", field, keyword.QueryString))
			}
		}
		tq.repr += strings.Join(keywords, " ")

		// Process the children.
		for _, child := range q.Children {
			tq.repr += t.Compile(child).String()
		}
		tq.repr += ")"
	} else if len(q.Operator) > 3 && q.Operator[0:3] == "adj" {
		tq.repr += " \""

		keywords := []string{}
		for _, keyword := range q.Keywords {
			for _, field := range keyword.Fields {
				keywords = append(keywords, fmt.Sprintf("%s:%s", field, keyword.QueryString))
			}
		}
		tq.repr += strings.Join(keywords, " ")

		// Process the children.
		for _, child := range q.Children {
			tq.repr += t.Compile(child).String()
		}

		distance := q.Operator[3:]
		tq.repr += fmt.Sprintf("\"~%s ", distance)
	} else {
		tq.repr = "("

		keywords := []string{}
		for _, keyword := range q.Keywords {
			for _, field := range keyword.Fields {
				keywords = append(keywords, fmt.Sprintf("%s:%s", field, keyword.QueryString))
			}
		}
		tq.repr += strings.Join(keywords, " ")

		// Process the children.
		for _, child := range q.Children {
			tq.repr += t.Compile(child).String()
		}
		tq.repr += ")"
	}

	return tq
}

func NewTerrierBackend() TerrierBackend {
	return TerrierBackend{}
}

func NewTerierQuery(repr string) TerrierQuery {
	return TerrierQuery{repr: repr}
}
