package transmute

import (
	"github.com/hscells/transmute/pipeline"
	"github.com/hscells/transmute/parser"
	"github.com/hscells/transmute/backend"
	"github.com/hscells/transmute/lexer"
)

var (
	Medline2Cqr = pipeline.NewPipeline(
		parser.NewMedlineParser(),
		backend.NewCQRBackend(),
		pipeline.TransmutePipelineOptions{
			LexOptions: lexer.LexOptions{
				FormatParenthesis: false,
			},
			RequiresLexing: true,
		})
	Pubmed2Cqr = pipeline.NewPipeline(
		parser.NewPubMedParser(),
		backend.NewCQRBackend(),
		pipeline.TransmutePipelineOptions{
			LexOptions: lexer.LexOptions{
				FormatParenthesis: true,
			},
			RequiresLexing: false,
		})
	Cqr2Medline = pipeline.NewPipeline(
		parser.NewCQRParser(),
		backend.NewMedlineBackend(),
		pipeline.TransmutePipelineOptions{
			LexOptions: lexer.LexOptions{
				FormatParenthesis: false,
			},
			RequiresLexing: false,
		})
)
