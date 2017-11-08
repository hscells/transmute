package parser

import (
	"github.com/hscells/transmute/lexer"
	"testing"
)

var (
	cqrQuery = `{
    "operator": "or",
    "children": [
        {
            "query": "MiniMental",
            "fields": [
                "title",
                "abstract"
            ],
            "options": {
                "exploded": false,
                "truncated": false
            }
        },
        {
            "query": "\"mini mental stat*\"",
            "fields": [
                "title",
                "abstract"
            ],
            "options": {
                "exploded": false,
                "truncated": false
            }
        },
        {
            "query": "MMSE*",
            "fields": [
                "title",
                "abstract"
            ],
            "options": {
                "exploded": false,
                "truncated": false
            }
        },
        {
            "query": "sMMSE",
            "fields": [
                "title",
                "abstract"
            ],
            "options": {
                "exploded": false,
                "truncated": false
            }
        },
        {
            "query": "Folstein*",
            "fields": [
                "title",
                "abstract"
            ],
            "options": {
                "exploded": false,
                "truncated": false
            }
        }
    ]
}`
)

func TestCQR(t *testing.T) {
	ast := lexer.Node{
		Value:     cqrQuery,
		Children:  nil,
		Operator:  "",
		Reference: 1,
	}
	queryRep := NewCQRParser().Parse(ast)

	//t.Log(queryRep)

	expected := 5
	got := len(queryRep.Terms())
	if expected != got {
		t.Fatalf("Expected %v terms, got %v", expected, got)
	}

}
