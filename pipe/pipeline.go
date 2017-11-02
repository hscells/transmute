package pipe

import (
	"github.com/hscells/transmute/backend"
	"github.com/hscells/transmute/lexer"
	"github.com/hscells/transmute/parser"
)

// Pipeline contains the information needed to execute a full compilation.
type Pipeline struct {
	Parser       parser.QueryParser
	Compiler     backend.Compiler
	FieldMapping map[string][]string
}

// Execute takes a pipeline and a query and will fully lex, parse, and compile the query.
func (p Pipeline) Execute(query string) (backend.BooleanQuery, error) {
	// Set the field mapping on the parser if it is defined separately in the pipeline.
	// Otherwise, the default field mapping will be used for the parser.
	if p.FieldMapping != nil || len(p.FieldMapping) > 0 {
		p.Parser.FieldMapping = p.FieldMapping
	}

	// Lex.
	ast, err := lexer.Lex(query)
	if err != nil {
		return nil, err
	}

	// Parse.
	boolQuery := p.Parser.Parse(ast)

	// Compile.
	return p.Compiler.Compile(boolQuery), nil

}
