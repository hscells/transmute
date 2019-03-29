package transmute

import (
	"github.com/hscells/cqr"
	"github.com/hscells/transmute/backend"
	"github.com/hscells/transmute/lexer"
	"github.com/hscells/transmute/parser"
	"github.com/hscells/transmute/pipeline"
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
			RequiresLexing:          false,
			AddRedundantParenthesis: false,
		})
	Cqr2Pubmed = pipeline.NewPipeline(
		parser.NewCQRParser(),
		backend.NewPubmedBackend(),
		pipeline.TransmutePipelineOptions{
			LexOptions: lexer.LexOptions{
				FormatParenthesis: false,
			},
			RequiresLexing:          false,
			AddRedundantParenthesis: false,
		})
)

func CompileMedline2Cqr(q string) (cqr.CommonQueryRepresentation, error) {
	bq, err := Medline2Cqr.Execute(q)
	if err != nil {
		return nil, err
	}

	repr, err := bq.Representation()
	if err != nil {
		return nil, err
	}

	return repr.(cqr.CommonQueryRepresentation), nil
}

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

func CompileCqr2PubMed(q cqr.CommonQueryRepresentation) (string, error) {
	s, err := backend.NewCQRQuery(q).String()
	if err != nil {
		return "", err
	}

	b, err := Cqr2Pubmed.Execute(s)
	if err != nil {
		return "", err
	}

	p, err := b.String()
	if err != nil {
		return "", err
	}

	return p, nil
}

func CompileCqr2Medline(q cqr.CommonQueryRepresentation) (string, error) {
	s, err := backend.NewCQRQuery(q).String()
	if err != nil {
		return "", err
	}

	b, err := Cqr2Medline.Execute(s)
	if err != nil {
		return "", err
	}

	p, err := b.String()
	if err != nil {
		return "", err
	}

	return p, nil
}

func CompileCqr2String(q cqr.CommonQueryRepresentation) (string, error) {
	return backend.NewCQRQuery(q).String()
}

func CompileString2Cqr(q string) (cqr.CommonQueryRepresentation, error) {
	p := pipeline.NewPipeline(
		parser.NewCQRParser(),
		backend.NewCQRBackend(),
		pipeline.TransmutePipelineOptions{
			LexOptions: lexer.LexOptions{
				FormatParenthesis: false,
			},
			AddRedundantParenthesis: false,
			RequiresLexing:          false,
		})

	b, err := p.Execute(q)
	if err != nil {
		return nil, err
	}
	i, err := b.Representation()
	if err != nil {
		return nil, err
	}
	return i.(cqr.CommonQueryRepresentation), nil
}
