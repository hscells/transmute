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
	case "or" :
		elasticSearchBooleanQuery.grouping = "should"
	case "not":
		elasticSearchBooleanQuery.grouping = "must_not"
	default:
		elasticSearchBooleanQuery.grouping = "must"
	}

	elasticSearchBooleanQuery.queries = queries
	return elasticSearchBooleanQuery
}

func traverseGroup(node map[string]interface{}, q ElasticSearchBooleanQuery) map[string]interface{} {
	// a group is a node in the tree
	group := map[string]interface{}{}

	// the children can either be queries (depth of 1) or other, nested boolean queries (depth of n)
	// https://github.com/golang/go/wiki/InterfaceSlice
	var groups []interface{} = make([]interface{}, len(q.queries) + len(q.children))
	subQuery := 0

	for i := range q.queries {
		// choose a multi_match or a match if the query has > 1 field associated with it or not
		query := map[string]interface{}{}
		var fields = q.queries[i].fields

		// TODO maybe it should default to _all if there are no fields?
		queryString := q.queries[i].queryString

		// now, we can have a general way of constructing the query
		if len(fields) > 1 {
			query = map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query": queryString,
					"fields": fields,
				},
			}
			// add the phrase type if there are spaces
			if strings.Contains(queryString, " ") {
				q := query["multi_match"].(map[string]interface{})
				q["type"] = "phrase"
				query["multi_match"] = q
			}
		} else if len(fields) > 0 {
			// one type of query is needed for matching phrases
			if strings.Contains(queryString, " ") {
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
	return node
}

func (q ElasticSearchBooleanQuery) Children() []ElasticSearchBooleanQuery {
	return q.children
}

func (q ElasticSearchBooleanQuery) String() string {
	b, _ := json.Marshal(traverseGroup(map[string]interface{}{}, q))
	return string(b)
}

func (q ElasticSearchBooleanQuery) StringPretty() string {
	b, _ := json.MarshalIndent(traverseGroup(map[string]interface{}{}, q), "", "    ")
	return string(b)
}