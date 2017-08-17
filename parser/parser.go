// Package parser parses the query strings inside a search strategy. This package represents the intermediate step
// between the lexer and ir.
package parser

import (
	"github.com/hscells/transmute/lexer"
	"github.com/hscells/transmute/ir"
)

type QueryTransformer interface {
	Transform(query string) ir.Keyword
}

type QueryParser struct {
	FieldMapping map[string][]string
	Parser       QueryTransformer
}

func (q QueryParser) Parse(ast lexer.Node) ir.BooleanQuery {
	var visit func(node lexer.Node, query ir.BooleanQuery) ir.BooleanQuery
	visit = func(node lexer.Node, query ir.BooleanQuery) ir.BooleanQuery {
		query.Operator = node.Operator
		for _, child := range node.Children {
			if len(child.Operator) == 0 {
				query.Keywords = append(query.Keywords, q.Parser.Transform(child.Value))
			} else {
				query.Children = append(query.Children, visit(child, ir.BooleanQuery{}))
			}
		}
		return query
	}

	return visit(ast, ir.BooleanQuery{})
}
