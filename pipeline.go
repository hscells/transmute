package main

import (
	"github.com/hscells/transmute/parser"
	"github.com/hscells/transmute/backend"
	"github.com/hscells/transmute/lexer"
)

// Pipeline contains the information needed to execute a full compilation.
type Pipeline struct {
	Parser parser.QueryParser
	Compiler backend.Compiler
}

// Execute takes a pipeline and a query and will fully lex, parse, and compile the query.
func Execute(pipeline Pipeline, query string) (backend.BooleanQuery, error) {

	// Lex.
	ast, err := lexer.Lex(query)
	if err != nil {
		return nil, err
	}

	// Parse.
	boolQuery := pipeline.Parser.Parse(ast)

	// Compile.
	return pipeline.Compiler.Compile(boolQuery), nil

}