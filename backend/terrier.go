package backend

import (
	"github.com/hscells/transmute/ir"
	"fmt"
	"strings"
)

// TerrierQuery is the transmute representation of terrier queries.
type TerrierQuery struct {
	repr string
}

// TerrierBackend is the terrier query compiler.
type TerrierBackend struct{}

// String returns a JSON-encoded representation of the cqr.
func (q TerrierQuery) String() (string, error) {
	return q.repr, nil
}

// StringPretty returns a pretty-printed JSON-encoded representation of the cqr.
func (q TerrierQuery) StringPretty() (string, error) {
	return q.String()
}

// Representation of a terrier query.
func (q TerrierQuery) Representation() (interface{}, error) {
	return q.String()
}

// Compile a terrier query.
func (t TerrierBackend) Compile(q ir.BooleanQuery) (BooleanQuery, error) {
	tq := TerrierQuery{}

	// Process the keywords.
	if q.Operator == "and" {
		tq.repr = "("
		var keywords []string
		for _, keyword := range q.Keywords {
			for _, field := range keyword.Fields {
				keywords = append(keywords, fmt.Sprintf("+%s:%s", field, keyword.QueryString))
			}
		}
		tq.repr += strings.Join(keywords, " ")

		// Process the children.
		for _, child := range q.Children {
			c, err := t.Compile(child)
			if err != nil {
				return nil, err
			}
			s, _ := c.String()
			tq.repr += s
		}
		tq.repr += ")"
	} else if len(q.Operator) > 3 && q.Operator[0:3] == "adj" {
		tq.repr += " \""

		var keywords []string
		for _, keyword := range q.Keywords {
			for _, field := range keyword.Fields {
				keywords = append(keywords, fmt.Sprintf("%s:%s", field, keyword.QueryString))
			}
		}
		tq.repr += strings.Join(keywords, " ")

		// Process the children.
		for _, child := range q.Children {
			c, err := t.Compile(child)
			if err != nil {
				return nil, err
			}
			s, _ := c.String()
			tq.repr += s
		}

		distance := q.Operator[3:]
		tq.repr += fmt.Sprintf("\"~%s ", distance)
	} else {
		tq.repr = "("

		var keywords []string
		for _, keyword := range q.Keywords {
			for _, field := range keyword.Fields {
				keywords = append(keywords, fmt.Sprintf("%s:%s", field, keyword.QueryString))
			}
		}
		tq.repr += strings.Join(keywords, " ")

		// Process the children.
		for _, child := range q.Children {
			c, err := t.Compile(child)
			if err != nil {
				return nil, err
			}
			s, _ := c.String()
			tq.repr += s
		}
		tq.repr += ")"
	}

	return tq, nil
}

func NewTerrierBackend() TerrierBackend {
	return TerrierBackend{}
}

func NewTerierQuery(repr string) TerrierQuery {
	return TerrierQuery{repr: repr}
}
