// Package backend contains translation code from the immediate representation into a concrete query usable by a search
// engine.
//
// Implementing a backend requires implementing both the BooleanQuery interface and the Compiler interface.
package backend

import "github.com/hscells/transmute/ir"

// BooleanQuery is an interface for handling the queries in a query language. The most important method is String(),
// which will output an appropriate query suitable for a search engine.
type BooleanQuery interface {
	Representation() interface{}
	String() string
	StringPretty() string
}

// Compiler is an interface which requires the implementation of a compiler.
type Compiler interface {
	// Compile will transform an immediate representation into the corresponding boolean query for the backend. This
	// is the reason both the backend and query interfaces must be implemented for this package.
	Compile(ir ir.BooleanQuery) BooleanQuery
}