package parser

import (
	"github.com/hscells/transmute/lexer"
	"testing"
	"github.com/hscells/transmute/pipeline"
	"github.com/hscells/transmute/backend"
	"go/parser"
)

var (
	pubmedQueryString = `(((\"Contraceptive Agents, Female\"[Mesh] OR \"Contraceptive Devices, Female\"[Mesh] OR contracept*[tiab]) AND (\"Body Weight\"[Mesh] OR weight[tiab] OR \"Body Mass Index\"[Mesh])) NOT (cancer*[ti] OR polycystic [ti] OR exercise [ti] OR physical activity[ti] OR postmenopaus*[ti]))`
)

func TestPubMed_BooleanQuery_Terms(t *testing.T) {
	ast, err := lexer.Lex(pubmedQueryString)
	if err != nil {
		t.Fatal(err)
	}
	queryRep := NewPubMedParser().Parse(ast)
	if err != nil {
		t.Fatal(err)
	}

	expected := 11
	got := len(queryRep.Terms())
	if expected != got {
		t.Fatalf("Expected %v terms, got %v", expected, got)
	}
}

func TestPubMed_BooleanQuery_Fields(t *testing.T) {
	ast, err := lexer.Lex(pubmedQueryString)
	if err != nil {
		t.Fatal(err)
	}
	queryRep := NewPubMedParser().Parse(ast)
	if err != nil {
		t.Fatal(err)
	}

	expected := 13
	got := len(queryRep.Fields())
	if expected != got {
		t.Fatalf("Expected %v fields, got %v", expected, got)
	}
}

func TestPubMed_BooleanQuery_FieldCount(t *testing.T) {
	ast, err := lexer.Lex(pubmedQueryString)
	if err != nil {
		t.Fatal(err)
	}
	queryRep := NewPubMedParser().Parse(ast)
	if err != nil {
		t.Fatal(err)
	}

	expected := 4
	got := queryRep.FieldCount()["mesh_headings"]
	if expected != got {
		t.Fatalf("Expected %v fields, got %v", expected, got)
	}
}
