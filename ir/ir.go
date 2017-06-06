// Package ir contains code relating to the immediate representation query structure of a search strategy.
package ir

type Keyword struct {
	Id          int          `json:"id"`
	QueryString string       `json:"query"`
	Fields      []string     `json:"fields"`
	Exploded    bool         `json:"exploded"`
	Truncated   bool         `json:"truncated"`
}

type BooleanQuery struct {
	Operator string           `json:"operator"`
	Keywords []Keyword        `json:"keywords"`
	Children []BooleanQuery   `json:"children"`
}

// New creates a new IRBooleanQuery
func New(operator string) BooleanQuery {
	return BooleanQuery{
		Operator: operator,
		Keywords: []Keyword{},
		Children: []BooleanQuery{},
	}
}

// AddKeyword adds a new keyword to a boolean query
func (b BooleanQuery) AddKeyword(keyword Keyword) {
	b.Keywords = append(b.Keywords, keyword)
}

// AddChild adds a new child to a boolean query
func (b BooleanQuery) AddChild(child BooleanQuery) {
	b.Children = append(b.Children, child)
}