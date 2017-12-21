package backend

import (
	"encoding/json"
	"github.com/hscells/transmute/ir"
)

// IrQuery is just a wrapper for a boolean query in immediate representation.
type IrQuery struct {
	ir ir.BooleanQuery
}

// IrBackend is the implementation for an immediate representation backend.
type IrBackend struct{}

// Representation is the immediate representation using a native data structure.
func (q IrQuery) Representation() (interface{}, error) {
	return q.ir, nil
}

// String returns a JSON-encoded representation of the immediate representation.
func (q IrQuery) String() (string, error) {
	b, err := json.Marshal(q.ir)
	return string(b), err
}

// StringPretty returns a pretty-printed JSON-encoded representation of the immediate representation.
func (q IrQuery) StringPretty() (string, error) {
	b, err := json.MarshalIndent(q.ir, "", "    ")
	return string(b), err
}

// Compile returns a new IrQuery with the ir embedded inside.
func (b IrBackend) Compile(query ir.BooleanQuery) (BooleanQuery, error) {
	return IrQuery{ir: query}, nil
}

// NewIrBackend returns an immediate representation compiler backend.
func NewIrBackend() IrBackend {
	return IrBackend{}
}
