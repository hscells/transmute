// Package lexer build a query tree from the search strategy.
package lexer

import (
	"strings"
	"regexp"
	"strconv"
)

var (
	numberRegex, _ = regexp.Compile("^[0-9]+$")
	prefixRegex, _ = regexp.Compile("^(or|and|not|OR|AND|NOT)/[0-9]+-[0-9]+$")
)

// Node contains the encoding of the query as a tree.
type Node struct {
	Value     string
	Reference int
	Operator  string
	Children  []Node
}

// ProcessInfixOperators replaces the references in an infix query with the actual query string.
func ProcessInfixOperators(queries map[int]string, operators string) (map[string]map[int]string, error) {
	extracted := map[int]string{}
	var operator string
	// We can be pretty sure that this will be correct.
	for i, token := range strings.Split(operators, " ") {
		// This is a bit of a hack but it does the job.
		if i%2 == 0 {
			reference, err := strconv.Atoi(token)
			if err != nil {
				return map[string]map[int]string{}, err
			}
			extracted[reference] = queries[reference-1]
		} else {
			operator = token
		}
	}
	return map[string]map[int]string{operator: extracted}, nil
}

// ProcessInfixOperators replaces the references in a prefix query with the actual query string.
func ProcessPrefixOperators(queries map[int]string, operator string) (map[string]map[int]string, error) {
	// Sort out the parts of the string.
	parts := strings.Split(operator, "/")
	op := parts[0]
	numbers := parts[1]
	numberParts := strings.Split(numbers, "-")
	// Grab the from and to numbers.
	from, err := strconv.Atoi(numberParts[0])
	if err != nil {
		return map[string]map[int]string{}, err
	}
	to, err := strconv.Atoi(numberParts[1])
	if err != nil {
		return map[string]map[int]string{}, err
	}
	// Generate the query mapping
	extracted := map[int]string{}
	for i := from - 1; i < to; i++ {
		extracted[i+1] = queries[i]
	}
	return map[string]map[int]string{op: extracted}, nil
}

// ExpandQuery takes a query that has been processed and expands it into a tree.
func ExpandQuery(query map[int]map[string]map[int]string) Node {
	var bottomReference int
	var operator string

	// Populate the top level node in the ast
	if len(query) == 1 {
		for k, v := range query {
			bottomReference = k
			for i := range v {
				operator = i
				break
			}
			break
		}
	} else {
		biggest := 0
		for k, v := range query {
			for i := range v {
				if k > biggest {
					operator = i
					biggest = k
				}
			}
		}
		bottomReference = biggest
	}
	// Walk down the ast children, adding nodes top down
	ast := Node{Reference: bottomReference, Operator: operator, Children: []Node{}}

	// This recursive function builds the tree recursively by adding nodes top down.
	var expand func(node Node, query map[int]map[string]map[int]string) Node
	expand = func(node Node, query map[int]map[string]map[int]string) Node {
		for k, v := range query[node.Reference][node.Operator] {
			// If we find a query in the top-level, process that.
			if innerQuery, ok := query[k]; ok {
				for operator := range innerQuery {
					n := Node{Reference: k, Operator: operator}
					//n.Children = append(n.Children, )
					node.Children = append(node.Children, expand(n, query))
				}
			} else {
				// Otherwise just append.
				node.Children = append(node.Children, Node{Reference: k, Value: v})
			}
		}
		return node
	}

	ast = expand(ast, query)

	return ast
}

// Lex creates the abstract syntax tree for the query. It will preprocess the query to try to normalise it. This
// function only creates the tree; it does not attempt to parse the individual lines in the query.
func Lex(query string) (Node, error) {
	query = PreProcess(query)

	// reference -> operator -> reference -> query_string
	depth1Query := map[int]map[string]map[int]string{}
	queries := map[int]string{}

	var err error
	// In the first pass, we create a depth-1 query structure.
	for reference, line := range strings.Split(query, "\n") {
		// First check if we are looking at an operator.

		if numberRegex.MatchString(strings.Split(line, " ")[0]) {
			// Assume we are looking at `N OP N OP N`.
			depth1Query[reference+1], err = ProcessInfixOperators(queries, line)
			if err != nil {
				return Node{}, err
			}
		} else if prefixRegex.MatchString(line) {
			// Assume we are looking at `OP/N-N
			depth1Query[reference+1], err = ProcessPrefixOperators(queries, line)
			if err != nil {
				return Node{}, err
			}
		}

		// We can be pretty sure that the string is for a query
		queries[reference] = line
	}

	var ast Node
	if len(depth1Query) == 0 {
		ast = Node{Value: queries[0], Reference: 1}
	} else {
		// In the second pass, we then parse a second time recursively to expand the inner queries at depth 1.
		ast = ExpandQuery(depth1Query)
	}

	return ast, nil
}
