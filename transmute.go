package main

import (
	"github.com/alexflint/go-arg"
	"log"
	"encoding/json"
	"github.com/hscells/transmute/parser"
	"os"
	"io/ioutil"
	"github.com/hscells/transmute/backend"
)

type args struct {
	Input        string `arg:"help:File containing a search strategy."`
	Output       string `arg:"help:File to output the transformed query to."`
	Parser       string `arg:"help:Which parser to use"`
	Backend      string `arg:"help:Which backend to use."`
	FieldMapping string `arg:"help:Load a field mapping json file."`
}

func (args) Version() string {
	return "transmute 29.Aug.2017"
}

func (args) Description() string {
	return `Pubmed/Medline query transpiler. Can read input from stdin and will output to stdout by default. See --help
for more details.
For further documentation see https://godoc.org/github.com/hscells/transmute.
To view the source or to contribute see https://github.com/hscells/transmute.`
}

func main() {
	var args args
	var query string
	inputFile := os.Stdin
	outputFile := os.Stdout

	pipeline := Pipeline{}

	// Parse the args into the struct
	arg.MustParse(&args)

	// Grab the input file (if defaults to stdin).
	if len(args.Input) > 0 {
		// Load the query
		fp, err := os.Open(args.Input)
		if err != nil {
			log.Fatal(err)
		}
		qb, err := ioutil.ReadAll(fp)
		if err != nil {
			log.Fatal(err)
		}
		query = string(qb)
	} else {
		data, err := ioutil.ReadAll(inputFile)
		if err != nil {
			log.Fatal(err)
		}
		query = string(data)
	}

	// Grab the output file (it defaults to stdout).
	if args.Output != "" {
		var err error
		outputFile, err = os.OpenFile(args.Output, os.O_WRONLY, 0)
		if os.IsNotExist(err) {
			outputFile, err = os.Create(args.Output)
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	if len(args.FieldMapping) > 0 {
		// Load the query
		fp, err := os.Open(args.Input)
		if err != nil {
			log.Fatal(err)
		}
		qb, err := ioutil.ReadAll(fp)
		if err != nil {
			log.Fatal(err)
		}

		var fieldMapping map[string][]string
		err = json.Unmarshal(qb, &fieldMapping)
		if err != nil {
			log.Fatal(err)
		}
		pipeline.FieldMapping = fieldMapping
	}

	// The list of available parsers.
	parsers := map[string]parser.QueryParser{
		"medline": parser.NewMedlineParser(),
	}

	// The list of available back-ends.
	compilers := map[string]backend.Compiler{
		"elasticsearch": backend.NewElasticsearchCompiler(),
		"ir": backend.NewIrBackend(),
	}

	// Grab the parser.
	if p, ok := parsers[args.Parser]; ok {
		pipeline.Parser = p
	} else {
		log.Fatalf("%v is not a valid parser", args.Parser)
	}

	// Grab the compiler.
	if c, ok := compilers[args.Backend]; ok {
		pipeline.Compiler = c
	} else {
		log.Fatalf("%v is not a valid backend", args.Backend)
	}

	// Execute the configured pipeline on the query.
	compiledQuery, err := Execute(pipeline, query)
	if err != nil {
		log.Fatal(err)
	}

	outputFile.WriteString(compiledQuery.StringPretty())

}
