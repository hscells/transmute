package parser

import (
	"github.com/hscells/transmute/ir"
	"github.com/prataprc/goparsec"
	"fmt"
)

type CochraneLibParser struct {
}

func (CochraneLibParser) TransformSingle(query string, mapping map[string][]string) ir.Keyword {
	panic("implement me")
}

func (CochraneLibParser) TransformNested(query string, mapping map[string][]string) ir.BooleanQuery {
	var array parsec.Parser

	ast := parsec.NewAST("CochraneLibrary", 100)

	id := parsec.And(
		func(ns []parsec.ParsecNode) parsec.ParsecNode {
			t := ns[0].(*parsec.Terminal)
			t.Value = `[` + t.Value + `]`
			return t
		},
		parsec.Token(`("[a-zA-Z0-9*? ]+"|[a-zA-Z0-9*?]+)`, "ID"),
	)

	operator := parsec.And(
		func(ns []parsec.ParsecNode) parsec.ParsecNode {
			t := ns[0].(*parsec.Terminal)
			t.Value = `$` + t.Value + `$`
			return t
		},
		parsec.Token(`(AND|OR|NOT){1}`, "OPERATOR"),
	)

	proximity := parsec.And(
		func(ns []parsec.ParsecNode) parsec.ParsecNode {
			t := ns[0].(*parsec.Terminal)
			t.Value = `$` + t.Value + `$`
			return t
		},
		parsec.Token(`(NEAR/[0-9]+|NEAR|NEXT){1}`, "PROXIMITY"),
	)

	fields := parsec.And(
		func(ns []parsec.ParsecNode) parsec.ParsecNode {
			t := ns[0].(*parsec.Terminal)
			t.Value = `~` + t.Value + `~`
			return t
		},
		parsec.Token(`:([a-z]+,|[a-z]+)+`, "FIELDS"),
	)

	opensqr := parsec.Atom(`(`, "OPENPAREN")
	closesqr := parsec.Atom(`)`, "CLOSEPAREN")
	space := ast.Maybe("space", nil, parsec.Atom(` `, "SPACE"))

	item := ast.OrdChoice("item", nil, operator, proximity, id, fields, &array)
	itemsep := ast.And("itemsep", nil, item, space)
	items := ast.Kleene("items", nil, itemsep, nil)
	array = ast.And("array", nil, opensqr, items, closesqr)

	s := parsec.NewScanner([]byte(query))
	node, _ := ast.Parsewith(array, s)
	fmt.Println(node.GetValue())

	return ir.BooleanQuery{}
}
