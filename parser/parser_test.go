package parser

import (
	"testing"
	"encoding/json"
	"github.com/hscells/transmute/lexer"
	"os"
	"io/ioutil"
	"log"
)

func TestParsr(t *testing.T) {
	q, _ := os.Open("../data/240")
	qp, _ := ioutil.ReadAll(q)

	ast, err := lexer.Lex(string(qp))
	if err != nil {
		panic(err)
	}

	parser := NewPubMedParser()
	query := parser.Parse(ast)

	//log.Println(string(qp))

	//p, _ := json.MarshalIndent(ast, "", "    ")
	//log.Println(string(p))

	p, _ := json.MarshalIndent(query, "", "    ")
	log.Println(string(p))

	//log.Println(backend.NewElasticSearchBackend().Compile(query).StringPretty())
}
