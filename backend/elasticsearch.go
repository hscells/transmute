// Package backend contains translation code from the immediate representation into a concrete query usable by a search
// engine.
//
// This file contains a decent reference implementation for compiling and transforming the IR into and ElasticSearch
// query.
package backend

import (
	"github.com/hscells/transmute/ir"
	"encoding/json"
	"strings"
	"github.com/pkg/errors"
	"fmt"
	"strconv"
)

type ElasticSearchQuery struct {
	queryString string
	fields      []string
}

type ElasticSearchBooleanQuery struct {
	queries  []ElasticSearchQuery
	grouping string
	children []ElasticSearchBooleanQuery
}

type ElasticSearchBackend struct{}

func NewElasticSearchBackend() ElasticSearchBackend {
	return ElasticSearchBackend{}
}

func (b ElasticSearchBackend) Compile(ir ir.BooleanQuery) ElasticSearchBooleanQuery {
	elasticSearchBooleanQuery := ElasticSearchBooleanQuery{}

	queries := []ElasticSearchQuery{}

	for _, keyword := range ir.Keywords {
		query := ElasticSearchQuery{}
		query.queryString = keyword.QueryString
		query.fields = keyword.Fields
		queries = append(queries, query)
	}

	for _, child := range ir.Children {
		elasticSearchBooleanQuery.children = append(elasticSearchBooleanQuery.children, b.Compile(child))
	}

	// This is really the only thing that differs from the IR; ElasticSearch has funny boolean operators
	switch ir.Operator {
	case "or":
		elasticSearchBooleanQuery.grouping = "should"
	case "not":
		elasticSearchBooleanQuery.grouping = "must_not"
	case "and":
		elasticSearchBooleanQuery.grouping = "filter"
	default:
		elasticSearchBooleanQuery.grouping = ir.Operator
	}

	elasticSearchBooleanQuery.queries = queries
	return elasticSearchBooleanQuery
}

func transformElasticsearchQuery(q ElasticSearchBooleanQuery) map[string]interface{} {
	return map[string]interface{}{
		"query": traverseGroup(map[string]interface{}{}, q),
	}
}

func traverseGroup(node map[string]interface{}, q ElasticSearchBooleanQuery) map[string]interface{} {
	// a group is a node in the tree
	group := map[string]interface{}{}

	// the children can either be queries (depth of 1) or other, nested boolean queries (depth of n)
	// https://github.com/golang/go/wiki/InterfaceSlice
	var groups []interface{} = make([]interface{}, len(q.queries)+len(q.children))
	subQuery := 0

	if len(q.grouping) > 3 && q.grouping[0:3] == "adj" {
		clauses := []interface{}{}
		for _, child := range q.children {
			// Panic if we egt anything other than a should.
			if child.grouping != "should" {
				panic(errors.New(fmt.Sprintf("unsupported operator for slop `%v`", child.grouping)))
			}
			// Create the clauses inside one side of the span
			innerClauses := []interface{}{}
			for _, query := range child.queries {
				if len(query.fields) != 1 {
					panic(errors.New(fmt.Sprintf("query `%v` has too many fields (%v)", query.queryString, query.fields)))
				}
				if strings.Contains(query.queryString, "*") || strings.Contains(query.queryString, "?") {
					innerClauses = append(innerClauses, map[string]interface{}{
						"span_multi": map[string]interface{}{
							"match": map[string]interface{}{
								"wildcard": map[string]interface{}{
									query.fields[0]: query.queryString,
								},
							},
						},
					})
				}
			}
			// Nest the inner clauses inside a span_or.
			clause := map[string]interface{}{
				"span_or": map[string]interface{}{
					"clauses": innerClauses,
				},
			}

			// Add this clause to the outer most span_near.
			clauses = append(clauses, clause)
		}

		// Extract the size of the adjacency (slop size)
		slopString := strings.Replace(q.grouping, "adj", "", -1)
		slopSize, err := strconv.Atoi(slopString)
		if err != nil {
			panic(err)
		}

		// Add both sides of the adjacency to the span.
		query := map[string]interface{}{
			"span_near": map[string]interface{}{
				"clauses": clauses,
				"slop":    slopSize,
			},
		}

		group = query
		node = group
	} else {
		for i := range q.queries {
			// choose a multi_match or a match if the query has > 1 field associated with it or not
			query := map[string]interface{}{}
			var fields = q.queries[i].fields

			// TODO maybe it should default to _all if there are no fields?
			queryString := q.queries[i].queryString

			// now, we can have a general way of constructing the query
			if len(fields) > 1 {
				if strings.ContainsAny(queryString, "*?") {
					queries := []interface{}{}

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
					// TODO this should be multiple should queries?
					query = map[string]interface{}{
						"multi_match": map[string]interface{}{
							"query":  queryString,
							"fields": fields,
						},
					}
					// add the phrase type if there are spaces
					if strings.Contains(queryString, " ") {
						q := query["multi_match"].(map[string]interface{})
						q["type"] = "phrase"
						query["multi_match"] = q
					}
				}
			} else if len(fields) > 0 {
				// Check to see if we first need to create a wildcard query.
				if strings.ContainsAny(queryString, "*?") {
					query = map[string]interface{}{
						"wildcard": map[string]interface{}{
							fields[0]: queryString,
						},
					}
				} else if strings.Contains(queryString, " ") {
					// one type of query is needed for matching phrases
					query = map[string]interface{}{
						"match_phrase": map[string]interface{}{
							fields[0]: queryString,
						},
					}
				} else {
					// otherwise we just use a regular match query
					query = map[string]interface{}{
						"match": map[string]interface{}{
							fields[0]: queryString,
						},
					}
				}
			}

			groups[subQuery] = query
			subQuery++
		}

		// and then the children
		for i := range q.children {
			// children are non-terminal so we descend down the tree
			groups[subQuery] = traverseGroup(map[string]interface{}{}, q.children[i])
			subQuery++
		}

		// finally, we have a layer to the tree, so return it upwards
		group[q.grouping] = groups
		group["disable_coord"] = true
		node["bool"] = group
	}

	return node
}

func (q ElasticSearchBooleanQuery) Children() []ElasticSearchBooleanQuery {
	return q.children
}

func (q ElasticSearchBooleanQuery) String() string {
	b, _ := json.Marshal(transformElasticsearchQuery(q))
	return string(b)
}

func (q ElasticSearchBooleanQuery) StringPretty() string {
	b, _ := json.MarshalIndent(transformElasticsearchQuery(q), "", "    ")
	return string(b)
}
