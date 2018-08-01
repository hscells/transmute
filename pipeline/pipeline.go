package pipeline

import (
	"github.com/hscells/transmute/backend"
	"github.com/hscells/transmute/lexer"
	"github.com/hscells/transmute/parser"
	"strings"
	"fmt"
)

// TransmutePipeline contains the information needed to execute a full compilation.
type TransmutePipeline struct {
	Parser   parser.QueryParser
	Compiler backend.Compiler
	Options  TransmutePipelineOptions
}

// TransmutePipelineOptions contains additional optional components relating to the pipeline.
type TransmutePipelineOptions struct {
	LexOptions              lexer.LexOptions
	FieldMapping            map[string][]string
	AddRedundantParenthesis bool
	RequiresLexing          bool
}

// NewPipeline creates a new transmute pipeline.
func NewPipeline(parser parser.QueryParser, compiler backend.Compiler, options TransmutePipelineOptions) TransmutePipeline {
	return TransmutePipeline{
		Parser:   parser,
		Compiler: compiler,
		Options:  options,
	}
}

// Execute takes a pipeline and a query and will fully lex, parse, and compile the query.
func (p TransmutePipeline) Execute(query string) (backend.BooleanQuery, error) {
	// Set the field mapping on the parser if it is defined separately in the pipeline.
	// Otherwise, the default field mapping will be used for the parser.
	if p.Options.FieldMapping != nil || len(p.Options.FieldMapping) > 0 {
		p.Parser.FieldMapping = p.Options.FieldMapping
	}

	// Lex.
	var ast lexer.Node
	var err error
	if p.Options.AddRedundantParenthesis {
		if strings.Count(query, "\n") == 0 {
			query = fmt.Sprintf("(%s)", query)
		}
	}
	if p.Options.RequiresLexing {
		ast, err = lexer.Lex(query, p.Options.LexOptions)
		if err != nil {
			return nil, err
		}
	} else {
		ast = lexer.Node{
			Value:     query,
			Children:  nil,
			Operator:  "",
			Reference: 1,
		}
	}

	// Parse.
	boolQuery := p.Parser.Parse(ast)

	// Compile.
	return p.Compiler.Compile(boolQuery)

}
