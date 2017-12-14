// Package backend contains translation code from the immediate representation into a concrete query usable by a search
// engine.
//
// This file contains a decent reference implementation for compiling and transforming the IR into and ElasticSearch
// query.
package backend

import (
	"encoding/json"
	"github.com/hscells/transmute/ir"
	"log"
	"strconv"
	"strings"
)

type ElasticsearchQuery struct {
	queryString string
	fields      []string
}

type ElasticsearchBooleanQuery struct {
	queries  []ElasticsearchQuery
	grouping string
	children []BooleanQuery
}

type ElasticsearchCompiler struct{}

// NewElasticsearchCompiler returns a new backend for compiling Elasticsearch queries.
func NewElasticsearchCompiler() ElasticsearchCompiler {
	return ElasticsearchCompiler{}
}

// Compile transforms an immediate representation of a query into an Elasticsearch query.
func (b ElasticsearchCompiler) Compile(ir ir.BooleanQuery) BooleanQuery {
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
			children = append(children, b.Compile(child))
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
			log.Println(queries)
			log.Println(children)
			log.Fatalln("a not query cannot have less than two children")
		}

		elasticSearchBooleanQuery.children = []BooleanQuery{rhsQuery, lhsQuery}
		elasticSearchBooleanQuery.grouping = "filter"

	} else {
		elasticSearchBooleanQuery.queries = queries

		for _, child := range ir.Children {
			elasticSearchBooleanQuery.children = append(elasticSearchBooleanQuery.children, b.Compile(child))
		}

		if len(elasticSearchBooleanQuery.grouping) == 0 {
			log.Fatalf("no operator was defined for an Elasticsearch query, context: %v", ir.Keywords)
		}
	}

	return elasticSearchBooleanQuery
}

// Representation is a wrapper for the traverseGroup function. This function should be used to transform
// the Elasticsearch ir into a valid Elasticsearch query.
func (q ElasticsearchBooleanQuery) Representation() interface{} {
	return map[string]interface{}{
		"query": q.traverseGroup(map[string]interface{}{}),
	}
}

// traverseGroup recursively transforms the Elasticsearch ir into a valid Elasticsearch query representable in JSON.
func (q ElasticsearchBooleanQuery) traverseGroup(node map[string]interface{}) map[string]interface{} {
	// a group is a node in the tree
	group := map[string]interface{}{}

	// the children can either be queries (depth of 1) or other, nested boolean queries (depth of n)
	// https://github.com/golang/go/wiki/InterfaceSlice
	var groups = make([]interface{}, len(q.queries)+len(q.children))
	subQuery := 0

	if q.grouping[0:3] == "adj" {
		clauses := map[string][]interface{}{}
		for _, child := range q.children {
			child := child.(ElasticsearchBooleanQuery)
			// Error if we get anything other than a should.
			if child.grouping != "should" {
				log.Fatalf("unsupported operator for slop `%v`\noffending query:\n%v", child.grouping, child.StringPretty())
			}
			// Create the clauses inside one side of the span.
			for _, query := range child.queries {
				for _, field := range query.fields {
					clauses[field] = append(clauses[field], query.createAdjacentClause(field)...)
				}
			}
		}

		// Now create the clauses for each of the queries at this level.
		for _, query := range q.queries {
			for _, field := range query.fields {
				clauses[field] = append(clauses[field], query.createAdjacentClause(field)...)
			}
		}

		// Extract the size of the adjacency (slop size)
		var slopSize int
		if len(q.grouping) > 3 {
			slopString := strings.Replace(q.grouping, "adj", "", -1)
			var err error
			slopSize, err = strconv.Atoi(slopString)
			if err != nil {
				panic(err)
			}
		} else {
			slopSize = 1
		}

		var query map[string]interface{}
		if len(clauses) == 1 {
			// Add both sides of the adjacency to the span.
			for _, clause := range clauses {
				query = map[string]interface{}{
					"span_near": map[string]interface{}{
						"clauses": clause,
						"slop":    slopSize,
					},
				}
			}
		} else if len(clauses) > 1 {
			queries := make([]interface{}, len(clauses))
			i := 0
			for _, clause := range clauses {
				queries[i] = map[string]interface{}{
					"span_near": map[string]interface{}{
						"clauses": clause,
						"slop":    slopSize,
					},
				}
				i++
			}
			outerSpan := map[string]interface{}{
				"bool": map[string]interface{}{
					"should": []interface{}{
						queries,
					},
				},
			}
			query = outerSpan
		}
		group = query
		node = group
	} else {
		for i := range q.queries {
			// Choose a multi_match or a match if the query has > 1 field associated with it or not.
			query := map[string]interface{}{}
			fields := q.queries[i].fields

			queryString := q.queries[i].queryString

			// Now, we can have a general way of constructing the query.
			if len(fields) > 1 {
				// Multiple fields, with a wildcard query string.
				if strings.ContainsAny(queryString, "*?") {
					var queries []interface{}

					for _, field := range fields {
						queries = append(queries, map[string]interface{}{
							"wildcard": map[string]interface{}{
								field: queryString,
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

						if strings.Contains(queryString, " ") {
							// One type of query is needed for matching phrases.
							queries = append(queries, map[string]interface{}{
								"match_phrase": map[string]interface{}{
									field: queryString,
								},
							})
						} else {
							// Otherwise we just use a regular match query.
							queries = append(queries, map[string]interface{}{
								"match": map[string]interface{}{
									field: queryString,
								},
							})
						}
					}
					query = map[string]interface{}{
						"bool": map[string]interface{}{
							"should": queries,
						},
					}
				}
			} else if len(fields) == 1 {
				// Check to see if we first need to create a wildcard query.
				if strings.ContainsAny(queryString, "*?") {
					query = map[string]interface{}{
						"wildcard": map[string]interface{}{
							fields[0]: queryString,
						},
					}
				} else if strings.Contains(queryString, " ") {
					// One type of query is needed for matching phrases.
					query = map[string]interface{}{
						"match_phrase": map[string]interface{}{
							fields[0]: queryString,
						},
					}
				} else {
					// Otherwise we just use a regular match query.
					query = map[string]interface{}{
						"match": map[string]interface{}{
							fields[0]: queryString,
						},
					}
				}
			} else {
				log.Fatalf("a query `%v` did not contain any fields", queryString)
			}

			groups[subQuery] = query
			subQuery++
		}

		// And then the children.
		for i := range q.children {
			// Children are non-terminal so we descend down the tree.
			groups[subQuery] = q.children[i].(ElasticsearchBooleanQuery).traverseGroup(map[string]interface{}{})
			subQuery++
		}

		// Finally, we have a layer to the tree, so return it upwards.
		group[q.grouping] = groups
		group["disable_coord"] = true
		node["bool"] = group
	}

	return node
}

// createAdjacentClause attempts to create an Elasticsearch version of the `adj` operator in Pubmed/Medline (slop).
func (q ElasticsearchQuery) createAdjacentClause(field string) []interface{} {
	var innerClauses []interface{}

	// Create the wildcard query.
	if strings.Contains(q.queryString, "*") || strings.Contains(q.queryString, "?") {
		innerClauses = append(innerClauses, map[string]interface{}{
			"span_multi": map[string]interface{}{
				"match": map[string]interface{}{
					"wildcard": map[string]interface{}{
						field: q.queryString,
					},
				},
			},
		})
	} else {
		// Create a term matching query.
		innerClauses = append(innerClauses, map[string]interface{}{
			"span_multi": map[string]interface{}{
				"match": map[string]interface{}{
					"prefix": map[string]interface{}{
						field: q.queryString,
					},
				},
			},
		})
	}
	return innerClauses
}

// String creates a machine-readable JSON Elasticsearch query.
func (q ElasticsearchBooleanQuery) String() string {
	b, _ := json.Marshal(q.Representation())
	return string(b)
}

// StringPretty creates a human-readable JSON Elasticsearch query.
func (q ElasticsearchBooleanQuery) StringPretty() string {
	b, _ := json.MarshalIndent(q.Representation(), "", "    ")
	return string(b)
}
