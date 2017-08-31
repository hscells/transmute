package backend

import (
	"github.com/hscells/transmute/ir"
	"encoding/json"
)

// IrQuery is just a wrapper for a boolean query in immediate representation.
type IrQuery struct {
	ir ir.BooleanQuery
}

// IrBackend is the implementation for an immediate representation backend.
type IrBackend struct{}

// Representation is the immediate representation using a native data structure.
func (q IrQuery) Representation() interface{} {
	return q.ir
}

// String returns a JSON-encoded representation of the immediate representation.
func (q IrQuery) String() string {
	b, _ := json.Marshal(q.ir)
	return string(b)
}

// StringPretty returns a pretty-printed JSON-encoded representation of the immediate representation.
func (q IrQuery) StringPretty() string {
	b, _ := json.MarshalIndent(q.ir, "", "    ")
	return string(b)
}

// Compile returns a new IrQuery with the ir embedded inside.
func (b IrBackend) Compile(query ir.BooleanQuery) BooleanQuery {
	return IrQuery{ir: query}
}

// NewIrBackend returns an immediate representation compiler backend.
func NewIrBackend() IrBackend {
	return IrBackend{}
}
