package transmute

import (
	"github.com/hscells/transmute/pipeline"
	"github.com/hscells/transmute/parser"
	"github.com/hscells/transmute/backend"
	"github.com/hscells/transmute/lexer"
	"github.com/hscells/cqr"
)

var (
	Medline2Cqr = pipeline.NewPipeline(
		parser.NewMedlineParser(),
		backend.NewCQRBackend(),
		pipeline.TransmutePipelineOptions{
			LexOptions: lexer.LexOptions{
				FormatParenthesis: false,
			},
			AddRedundantParenthesis: true,
			RequiresLexing:          true,
		})
	Pubmed2Cqr = pipeline.NewPipeline(
		parser.NewPubMedParser(),
		backend.NewCQRBackend(),
		pipeline.TransmutePipelineOptions{
			LexOptions: lexer.LexOptions{
				FormatParenthesis: true,
			},
			AddRedundantParenthesis: true,
			RequiresLexing:          false,
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
	Cqr2Pubmed = pipeline.NewPipeline(
		parser.NewCQRParser(),
		backend.NewPubmedBackend(),
		pipeline.TransmutePipelineOptions{
			LexOptions: lexer.LexOptions{
				FormatParenthesis: false,
			},
			RequiresLexing: false,
		})
)

func CompilePubmed2Cqr(q string) (cqr.CommonQueryRepresentation, error) {
	bq, err := Pubmed2Cqr.Execute(q)
	if err != nil {
		return nil, err
	}

	repr, err := bq.Representation()
	if err != nil {
		return nil, err
	}

	return repr.(cqr.CommonQueryRepresentation), nil
}
