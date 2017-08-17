package parser

type Node struct {
	// Reference to a pattern in the grammar.
	Reference string `json:"reference"`

	// Transitions nodes in the tree.
	Transitions []Node    `json:"transitions"`
}

type Grammar struct {
	// How should fields in the queries map to ir fields?
	Fields map[string][]string `json:"fields"`

	// Mapping of reference to patterns to compile.
	Patterns map[string]string `json:"patterns"`

	// The abstract syntax tree of the query.
	AST Node `json:"ast"`
}
