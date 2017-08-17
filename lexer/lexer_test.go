package lexer

import (
	"testing"
	"io/ioutil"
	"os"
	"encoding/json"
	"log"
)

func TestLexPubmedQuery(t *testing.T) {
	q, _ := os.Open("../data/240")
	qp, _ := ioutil.ReadAll(q)

	ast, err := Lex(string(qp))
	if err != nil {
		panic(err)
	}

	p, _ := json.MarshalIndent(ast, "", "    ")
	log.Println(string(p))
}
