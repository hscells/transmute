package lexer

import (
	"testing"
	"io/ioutil"
	"os"
)

func TestLexPubmedQuery(t *testing.T) {
	q, _ := os.Open("../data/454")
	qp, _ := ioutil.ReadAll(q)

	Lex(string(qp))

}
