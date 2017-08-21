// Package parser parses the query strings inside a search strategy. This package represents the intermediate step
// between the lexer and ir.
package parser

import (
	"github.com/hscells/transmute/lexer"
	"github.com/hscells/transmute/ir"
)

type QueryTransformer interface {
	TransformSingle(query string) ir.Keyword
	TransformNested(query string) ir.BooleanQuery
}

type QueryParser struct {
	FieldMapping map[string][]string
	Parser       QueryTransformer
}

func (q QueryParser) Parse(ast lexer.Node) ir.BooleanQuery {
	if ast.Children == nil && ast.Reference == 1 {
		return q.Parser.TransformNested(ast.Value)
	}


	var visit func(node lexer.Node, query ir.BooleanQuery) ir.BooleanQuery
	visit = func(node lexer.Node, query ir.BooleanQuery) ir.BooleanQuery {
		query.Operator = node.Operator
		for _, child := range node.Children {
			if len(child.Operator) == 0 {
				// Nested query.
				if child.Value[0] == '(' {
					query.Children = append(query.Children, q.Parser.TransformNested(child.Value))
				} else {
					// Regular line of a query.
					query.Keywords = append(query.Keywords, q.Parser.TransformSingle(child.Value))
				}
			} else {
				query.Children = append(query.Children, visit(child, ir.BooleanQuery{}))
			}
		}
		return query
	}

	return visit(ast, ir.BooleanQuery{})
}
