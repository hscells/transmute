package backend

import (
	"github.com/hscells/transmute/ir"
	"encoding/json"
)

type IrQuery struct {
	ir ir.BooleanQuery
}

type IrBackend struct {}

func (q IrQuery) Representation() interface{} {
	return q.ir
}

func (q IrQuery) String() string {
	b, _ := json.Marshal(q.ir)
	return string(b)
}

func (q IrQuery) StringPretty() string {
	b, _ := json.MarshalIndent(q.ir, "", "    ")
	return string(b)
}

func (b IrBackend) Compile(query ir.BooleanQuery) BooleanQuery {
	return IrQuery{ir: query}
}

func NewIrBackend() IrBackend {
	return IrBackend{}
}