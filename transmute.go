package main

import (
	"github.com/alexflint/go-arg"
	"log"
	"encoding/json"
	"github.com/hscells/transmute/parser"
	"os"
	"io/ioutil"
)

type args struct {
	Input          string `arg:"help:File containing a search strategy."`
	Output         string `arg:"help:File to output the transformed query to."`
	StartsAfter    string `arg:"help:Character the keywords in a search strategy start after."`
	FieldSeparator string `arg:"help:Character the separates a keyword from the field used to search on."`
}

func (args) Version() string {
	return "transmute 0.0.1"
}

func (args) Description() string {
	return "Pubmed/Medline query transpiler."
}

func main() {
	var args args
	var query string
	inputFile := os.Stdin
	outputFile := os.Stdout
	startsAfter := rune(0)
	fieldSeparator := rune('.')

	// Specify default values
	args.StartsAfter = " "
	args.FieldSeparator = "."

	// Parse the args into the struct
	arg.MustParse(&args)

	// Grab the input file (if defaults to stdin).
	if args.Input != "" {
		// Load the query
		query = parser.Load(args.Input)
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

	// Override the default values
	if len(args.StartsAfter) > 0 {
		startsAfter = rune(args.StartsAfter[0])
	}
	if len(args.FieldSeparator) > 0 {
		fieldSeparator = rune(args.FieldSeparator[0])
	}

	// Parse the query into our own format
	ir_query := parser.Parse(query, startsAfter, fieldSeparator)

	// format the output
	d, err := json.MarshalIndent(ir_query, "", "    ")
	if err != nil {
		panic(err)
	}

	outputFile.Write(d)
}
