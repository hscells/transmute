// Package parser parses the query strings inside a search strategy. This package represents the intermediate step
// between the lexer and ir.
package parser

import (
	"github.com/hscells/transmute/ir"
	"github.com/hscells/transmute/lexer"
)

// QueryTransformer must be implemented to parse queries.
type QueryTransformer interface {
	// TransformSingle transforms a single query string.
	TransformSingle(query string, mapping map[string][]string) ir.Keyword

	// TransformNested transforms a nested query - queries that start with a `(`.
	TransformNested(query string, mapping map[string][]string) ir.BooleanQuery
}

// QueryParser represents the full implementation of a query parser.
type QueryParser struct {
	// FieldMapping determines how fields are mapped for a query.
	FieldMapping map[string][]string

	// Parser is an implemented QueryTransformer.
	Parser QueryTransformer
}

// Parse takes an AST created from lexing a query and parses each node in it. It uses the TransformNested and
// TransformSingle functions defined by the Parser and the Field mapping to create an immediate representation tree.
func (q QueryParser) Parse(ast lexer.Node) ir.BooleanQuery {
	if ast.Children == nil && ast.Reference == 1 {
		return q.Parser.TransformNested(ast.Value, q.FieldMapping)
	}

	var visit func(node lexer.Node, query ir.BooleanQuery) ir.BooleanQuery
	visit = func(node lexer.Node, query ir.BooleanQuery) ir.BooleanQuery {
		query.Operator = node.Operator
		for _, child := range node.Children {
			if len(child.Operator) == 0 {
				// Nested query.
				if child.Value[0] == '(' {
					query.Children = append(query.Children, q.Parser.TransformNested(child.Value, q.FieldMapping))
				} else {
					// Regular line of a query.
					query.Keywords = append(query.Keywords, q.Parser.TransformSingle(child.Value, q.FieldMapping))
				}
			} else {
				query.Children = append(query.Children, visit(child, ir.BooleanQuery{}))
			}
		}
		return query
	}

	return visit(ast, ir.BooleanQuery{})
}
