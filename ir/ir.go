// Package ir contains code relating to the immediate representation query structure of a search strategy.
package ir

// Keyword represents a single string inside a search strategy. When these are reported, however, a keyword not only
// contains the phrase to search, but the fields in the database to search, how it is truncated, and if it is a mesh
// term, if the term has been exploded.
type Keyword struct {
	QueryString string   `json:"query"`
	Fields      []string `json:"fields"`
	Exploded    bool     `json:"exploded"`
	Truncated   bool     `json:"truncated"`
}

// BooleanQuery is the immediate representation of a boolean query for a search engine. This representation groups a
// list of keywords by a single operator, much like prefix notation. To combine operators, they can be added as children
// to a query. This means that there is no ambiguity to a query.
type BooleanQuery struct {
	// A boolean operator (e.g. "and", "or", "not")
	Operator string `json:"operator"`
	// A list of Keywords that appear as queries grouped by the operator
	Keywords []Keyword `json:"keywords"`
	// Any sub-queries, or children of the current query
	Children []BooleanQuery `json:"children"`
}
