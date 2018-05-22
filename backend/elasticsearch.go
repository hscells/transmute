// Package backend contains translation code from the immediate representation into a concrete query usable by a search
// engine.
//
// This file contains a decent reference implementation for compiling and transforming the IR into and ElasticSearch
// query.
package backend

import (
	"encoding/json"
	"github.com/hscells/transmute/ir"
	"strconv"
	"strings"
	"github.com/hscells/meshexp"
	"fmt"
	"github.com/pkg/errors"
)

// ElasticsearchQuery is the transmute representation of an Elasticsearch keyword.
type ElasticsearchQuery struct {
	queryString string
	fields      []string
}

// ElasticsearchBooleanQuery is the transmute representation of an Elasticsearch query.
type ElasticsearchBooleanQuery struct {
	queries  []ElasticsearchQuery
	grouping string
	children []BooleanQuery
}

// ElasticsearchCompiler is a compiler for Elasticsearch queries.
type ElasticsearchCompiler struct {
	tree *meshexp.MeSHTree
}

// m is a shorthand type for constructing large Elasticsearch queries.
type m map[string]interface{}

// NewElasticsearchCompiler returns a new backend for compiling Elasticsearch queries.
func NewElasticsearchCompiler() ElasticsearchCompiler {
	tree, err := meshexp.Default()
	if err != nil {
		panic(err)
	}

	return ElasticsearchCompiler{
		tree: tree,
	}
}

// Compile transforms an immediate representation of a query into an Elasticsearch query.
func (b ElasticsearchCompiler) Compile(ir ir.BooleanQuery) (BooleanQuery, error) {
	elasticSearchBooleanQuery := ElasticsearchBooleanQuery{}

	var queries []ElasticsearchQuery

	// This is really the only thing that differs from the IR; Elasticsearch has funny boolean operators.
	switch ir.Operator {
	case "or", "OR":
		elasticSearchBooleanQuery.grouping = "should"
	case "not", "NOT":
		elasticSearchBooleanQuery.grouping = "must_not"
	case "and", "AND":
		elasticSearchBooleanQuery.grouping = "filter"
	default:
		elasticSearchBooleanQuery.grouping = ir.Operator
	}

	for _, keyword := range ir.Keywords {
		query := ElasticsearchQuery{}
		query.queryString = keyword.QueryString
		query.fields = keyword.Fields
		queries = append(queries, query)

		if keyword.Exploded {
			terms := b.tree.Explode(query.queryString)
			for _, term := range terms {
				queries = append(queries, ElasticsearchQuery{
					queryString: term,
					fields:      keyword.Fields,
				})
			}
		}

	}

	if elasticSearchBooleanQuery.grouping == "must_not" {
		lhsQuery := ElasticsearchBooleanQuery{
			grouping: "must_not",
		}
		rhsQuery := ElasticsearchBooleanQuery{
			grouping: "filter",
		}

		var children []BooleanQuery
		for _, child := range ir.Children {
			c, err := b.Compile(child)
			if err != nil {
				return nil, err
			}
			children = append(children, c)
		}

		if len(children) > 1 && len(queries) == 0 {
			rhsQuery.children = []BooleanQuery{children[0]}
			lhsQuery.children = children[1:]
		} else if len(queries) > 1 && len(children) == 0 {
			rhsQuery.queries = []ElasticsearchQuery{queries[0]}
			lhsQuery.queries = queries[1:]
		} else if len(queries) == 1 && len(children) > 0 {
			rhsQuery.queries = queries
			lhsQuery.children = children
		} else {
			return nil, errors.New(fmt.Sprintf("a not query cannot have less than two children:\n%v\n%v", queries, children))
		}

		elasticSearchBooleanQuery.children = []BooleanQuery{rhsQuery, lhsQuery}
		elasticSearchBooleanQuery.grouping = "filter"

	} else {
		elasticSearchBooleanQuery.queries = queries

		//fmt.Println(len(ir.Keywords), len(ir.Children))
		if (len(ir.Keywords) == 0 || ir.Keywords == nil) && len(ir.Children) == 1 {
			c, err := b.Compile(ir.Children[0])
			if err != nil {
				return nil, err
			}
			elasticSearchBooleanQuery = c.(ElasticsearchBooleanQuery)
		} else {
			for _, child := range ir.Children {
				c, err := b.Compile(child)
				if err != nil {
					return nil, err
				}
				elasticSearchBooleanQuery.children = append(elasticSearchBooleanQuery.children, c)
			}
		}

		if len(elasticSearchBooleanQuery.queries) > 0 && len(elasticSearchBooleanQuery.grouping) == 0 {
			return nil, errors.New(fmt.Sprintf("no operator was defined for an Elasticsearch query, context: %v", elasticSearchBooleanQuery.queries))
		}
	}

	return elasticSearchBooleanQuery, nil
}

// Representation is a wrapper for the traverseGroup function. This function should be used to transform
// the Elasticsearch ir into a valid Elasticsearch query.
func (q ElasticsearchBooleanQuery) Representation() (interface{}, error) {
	f, err := q.traverseGroup(m{})
	if err != nil {
		return nil, err
	}
	return m{
		"query": m{
			"constant_score": m{
				"filter": f,
			},
		},
	}, nil
}

// traverseGroup recursively transforms the Elasticsearch ir into a valid Elasticsearch query representable in JSON.
func (q ElasticsearchBooleanQuery) traverseGroup(node m) (m, error) {
	// a group is a node in the tree
	group := m{}

	// the children can either be queries (depth of 1) or other, nested boolean queries (depth of n)
	// https://github.com/golang/go/wiki/InterfaceSlice
	var groups = make([]interface{}, len(q.queries)+len(q.children))
	subQuery := 0

	if len(q.grouping) >= 3 && q.grouping[0:3] == "adj" {
		adjClauses := map[string][]interface{}{}
		nesClauses := map[string][]interface{}{}
		var clauses []interface{}

		// Extract the size of the adjacency (slop size)
		slopSize := 0
		if len(q.grouping) > 3 {
			slopString := strings.Replace(q.grouping, "adj", "", -1)
			var err error
			slopSize, err = strconv.Atoi(slopString)
			if err != nil {
				panic(err)
			}
		}

		// Now create the clauses for each of the queries at this level.
		for _, query := range q.queries {
			for _, field := range query.fields {
				adjClauses[field] = append(adjClauses[field], query.createAdjacentClause(field))
			}
		}

		for _, child := range q.children {
			child := child.(ElasticsearchBooleanQuery)
			// Error if we get anything other than a should.
			if child.grouping != "should" {
				s, err := child.StringPretty()
				if err != nil {
					return nil, errors.New(fmt.Sprintf("unsupported operator for slop `%v` (can't show query)", child.grouping, ))
				}
				return nil, errors.New(fmt.Sprintf("unsupported operator for slop `%v`\noffending query:\n%v", child.grouping, s))
			}
			// Create the clauses inside one side of the span.
			for _, query := range child.queries {
				for _, field := range query.fields {
					c := m{
						"span_near": m{
							"clauses":  append(adjClauses[field], query.createAdjacentClause(field)),
							"slop":     slopSize,
							"in_order": false,
						},
					}
					clauses = append(clauses, c)
					nesClauses[field] = append(nesClauses[field], c)
				}
			}
		}

		var query map[string]interface{}
		if len(clauses) == 0 { // There were no "nested" clauses, only terms.
			var ac []interface{}
			for _, q := range adjClauses {
				ac = append(ac, m{
					"span_near": m{
						"clauses":  q,
						"slop":     slopSize,
						"in_order": false,
					},
				})
			}
			query = m{
				"bool": m{
					"should": ac,
				},
			}
		} else if len(clauses) > 0 && len(adjClauses) == 0 { // only "nested" clauses, no terms.
			var ac []interface{}
			for _, q := range nesClauses {
				ac = append(ac, m{
					"span_near": m{
						"clauses":  q,
						"slop":     slopSize,
						"in_order": false,
					},
				})
			}
			query = m{
				"bool": m{
					"should": ac,
				},
			}
		} else { // mixture of "nested" clauses and terms.
			query = m{
				"bool": m{
					"should": []interface{}{
						clauses,
					},
				},
			}
		}

		//if len(clauses) == 1 {
		//	// Add both sides of the adjacency to the span.
		//	for _, clause := range clauses {
		//		query = m{
		//			"span_near": m{
		//				"clauses":  clause,
		//				"slop":     slopSize,
		//				"in_order": false,
		//			},
		//		}
		//	}
		//} else if len(clauses) > 1 {
		//	queries := make([]interface{}, len(clauses))
		//	i := 0
		//	for _, clause := range clauses {
		//		queries[i] = m{
		//			"span_near": m{
		//				"clauses":  clause,
		//				"slop":     slopSize,
		//				"in_order": false,
		//			},
		//		}
		//		i++
		//	}
		//	outerSpan := m{
		//		"bool": m{
		//			"should": []interface{}{
		//				queries,
		//			},
		//		},
		//	}
		//	query = outerSpan
		//}
		group = query
		node = group
	} else {
		for i := range q.queries {
			// Choose a multi_match or a match if the query has > 1 field associated with it or not.
			query := map[string]interface{}{}
			fields := q.queries[i].fields

			queryString := q.queries[i].queryString

			matchType := "match"
			if strings.ContainsRune(queryString, ' ') {
				matchType = "match_phrase"
			}

			// Now, we can have a general way of constructing the query.
			if len(fields) > 1 {
				// Multiple fields, with a wildcard query string.
				if strings.ContainsAny(queryString, "*?~") {
					var queries []interface{}
					/*
					{
              			"query_string": {
                			"query": "text.stemmed:gonadotrop?in releasing hormone agonist*",
                			"analyze_wildcard": true,
               	 			"split_on_whitespace" : false
              			}
					}
					 */
					for _, field := range fields {
						queries = append(queries, m{
							"query_string": m{
								"query":               fmt.Sprintf("%v:%v", field, queryString),
								"analyze_wildcard":    true,
								"split_on_whitespace": false,
							},
						})
					}

					query = map[string]interface{}{
						"bool": map[string]interface{}{
							"should": queries,
						},
					}
				} else {
					var queries []interface{}
					// Multiple fields, with a regular query string.
					for _, field := range fields {
						if strings.ContainsAny(queryString, "*?~") {
							queries = append(queries, m{
								"query_string": m{
									"query":               fmt.Sprintf("%v:%v", field, queryString),
									"analyze_wildcard":    true,
									"split_on_whitespace": false,
								},
							})
						} else {
							// Otherwise we just use a regular match query.
							queries = append(queries, m{
								matchType: m{
									field: queryString,
								},
							})
						}
					}
					query = m{
						"bool": m{
							"should": queries,
						},
					}
				}
			} else if len(fields) == 1 {
				// Check to see if we first need to create a wildcard query.
				if strings.ContainsAny(queryString, "*?") {
					query = m{
						"query_string": m{
							"query":               fmt.Sprintf("%v:%v", fields[0], queryString),
							"analyze_wildcard":    true,
							"split_on_whitespace": false,
						},
					}
				} else {
					// Otherwise we just use a regular match query.
					query = m{
						matchType: m{
							fields[0]: queryString,
						},
					}
				}
			} else {
				return nil, errors.New(fmt.Sprintf("a query `%v` did not contain any fields", queryString))
			}

			groups[subQuery] = query
			subQuery++
		}

		// And then the children.
		for i := range q.children {
			// Children are non-terminal so we descend down the tree.
			g, err := q.children[i].(ElasticsearchBooleanQuery).traverseGroup(m{})
			if err != nil {
				return nil, err
			}
			groups[subQuery] = g
			subQuery++
		}

		// Finally, we have a layer to the tree, so return it upwards.
		group[q.grouping] = groups
		group["disable_coord"] = true
		node["bool"] = group
	}

	return node, nil
}

// createAdjacentClause attempts to create an Elasticsearch version of the `adj` operator in Pubmed/Medline (slop).
func (q ElasticsearchQuery) createAdjacentClause(field string) map[string]interface{} {
	innerClauses := make(map[string]interface{})

	// Create the wildcard query.
	if strings.ContainsAny(q.queryString, "*?$~") {
		q.queryString = strings.Replace(q.queryString, "?", "*", -1)
		q.queryString = strings.Replace(q.queryString, "$", "*", -1)
		if strings.Contains(q.queryString, " ") {
			var spanTerms []m
			for _, term := range strings.Split(q.queryString, " ") {
				if strings.ContainsAny(term, "*?$~") {
					spanTerms = append(spanTerms, m{
						"span_multi": m{
							"match": m{
								"wildcard": m{
									field: term,
								},
							},
						},
					})
				} else {
					spanTerms = append(spanTerms, m{
						"span_term": m{
							field: term,
						},
					})
				}
			}
			innerClauses = m{
				"span_near": m{
					"clauses":  spanTerms,
					"in_order": true,
					"slop":     1,
				},
			}
		} else {
			innerClauses = m{
				"span_multi": m{
					"match": m{
						"wildcard": m{
							field: q.queryString,
						},
					},
				},
			}
		}
	} else if strings.Contains(q.queryString, " ") {
		var spanTerms []m
		for _, term := range strings.Split(q.queryString, " ") {
			spanTerms = append(spanTerms, m{
				"span_term": m{
					field: term,
				},
			})
		}
		innerClauses = m{
			"span_near": m{
				"clauses":  spanTerms,
				"in_order": true,
				"slop":     1,
			},
		}
	} else {
		// Create a term matching query.
		innerClauses = m{
			"span_multi": m{
				"match": m{
					"prefix": m{
						field: q.queryString,
					},
				},
			},
		}
	}
	return innerClauses
}

// String creates a machine-readable JSON Elasticsearch query.
func (q ElasticsearchBooleanQuery) String() (string, error) {
	r, err := q.Representation()
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(r)
	return string(b), err
}

// StringPretty creates a human-readable JSON Elasticsearch query.
func (q ElasticsearchBooleanQuery) StringPretty() (string, error) {
	r, err := q.Representation()
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(r, "", "  ")
	return string(b), err
}
