package main

import (
	"github.com/alexflint/go-arg"
	"log"
	"encoding/json"
	"github.com/hscells/transmute/parser"
	"os"
	"io/ioutil"
	"github.com/hscells/transmute/backend"
	"fmt"
	"github.com/hscells/transmute/lexer"
	"github.com/hscells/transmute/ir"
)

type args struct {
	Input   string `arg:"help:File containing a search strategy."`
	Output  string `arg:"help:File to output the transformed query to."`
	Parser  string `arg:"help:Which parser to use (medline)"`
	Backend string `arg:"help:Which backend to use (ir/elasticsearch)."`
}

func (args) Version() string {
	return "transmute 21.Aug.2017"
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

	// Specify default values
	args.Parser = "medline"
	args.Backend = "elasticsearch"

	// Parse the args into the struct
	arg.MustParse(&args)

	// Make sure the backend exists
	if args.Backend != "ir" && args.Backend != "elasticsearch" {
		fmt.Println(fmt.Sprintf("%v is not a valid backend. See `transmute --help` for details.", args.Backend))
		os.Exit(1)
	}

	// Grab the input file (if defaults to stdin).
	if len(args.Input) > 0 {
		// Load the query
		fp, err := os.Open(args.Input)
		if err != nil {
			panic(err)
		}
		qb, err := ioutil.ReadAll(fp)
		if err != nil {
			panic(err)
		}
		query = string(qb)
	} else {
		data, err := ioutil.ReadAll(inputFile)
		if err != nil {
			log.Panicln(err)
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
			log.Panicln(err)
		}
	}

	ast, err := lexer.Lex(query)
	if err != nil {
		panic(err)
	}

	var immediate ir.BooleanQuery
	switch args.Backend {
	case "medline":
	default:
		immediate = parser.NewMedlineParser().Parse(ast)
	}

	// Output the query
	switch args.Backend {
	case "ir":
		// format the output
		d, err := json.MarshalIndent(immediate, "", "    ")
		if err != nil {
			panic(err)
		}
		outputFile.Write(d)
	case "elasticsearch":
		outputFile.WriteString(backend.NewElasticSearchBackend().Compile(immediate).StringPretty())
	}
}
