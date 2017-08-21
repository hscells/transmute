package parser

import (
	"testing"
	"github.com/hscells/transmute/lexer"
	"os"
	"io/ioutil"
	"log"
	"github.com/hscells/transmute/backend"
	"encoding/json"
)

func TestParsr(t *testing.T) {
	q, _ := os.Open("../data/445")
	qp, _ := ioutil.ReadAll(q)

	ast, err := lexer.Lex(string(qp))
	if err != nil {
		panic(err)
	}

	parser := NewMedlineParser()
	query := parser.Parse(ast)

	log.Println(string(qp))

	p, _ := json.MarshalIndent(ast, "", "    ")
	log.Println(string(p))

	p, _ = json.MarshalIndent(query, "", "    ")
	log.Println(string(p))

	log.Println(backend.NewElasticSearchBackend().Compile(query).StringPretty())
}
