package lexer

import (
	"testing"
)

var (
	medlineQueryString = `1. exp Sleep Apnea Syndromes/
2. (sleep$ adj3 (apnea$ or apnoea$)).mp.
3. (hypopnoea$ or hypopnoea$).mp.
4. OSA.mp.
5. SHS.mp.
6. OSAHS.mp.
7. or/1-6`
	pubmedQueryString = `(((\"Contraceptive Agents, Female\"[Mesh] OR \"Contraceptive Devices, Female\"[Mesh] OR contracept*[tiab]) AND (\"Body Weight\"[Mesh] OR weight[tiab] OR \"Body Mass Index\"[Mesh])) NOT (cancer*[ti] OR polycystic [ti] OR exercise [ti] OR physical activity[ti] OR postmenopaus*[ti]))`
)

func Test_Lex_MedlineQuery(t *testing.T) {
	ast, err := Lex(string(medlineQueryString))
	if err != nil {
		panic(err)
	}

	expected := 6
	got := len(ast.Children)
	if expected != got {
		t.Fatalf("expected %v children, got %v", expected, got)
	}
}

func Test_Lex_PubMedQuery(t *testing.T) {
	ast, err := Lex(string(pubmedQueryString))
	if err != nil {
		panic(err)
	}

	expected := `(((\"Contraceptive Agents, Female\"[Mesh] OR \"Contraceptive Devices, Female\"[Mesh] OR contracept*[tiab]) AND (\"Body Weight\"[Mesh] OR weight[tiab] OR \"Body Mass Index\"[Mesh])) NOT (cancer*[ti] OR polycystic [ti] OR exercise [ti] OR physical activity[ti] OR postmenopaus*[ti]))`
	got := ast.Value
	if expected != got {
		t.Fatalf("expected %v children, got %v", expected, got)
	}
}
