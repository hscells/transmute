package backend

import (
	"github.com/hscells/transmute/ir"
	"fmt"
	"github.com/hscells/transmute/parser"
	"strings"
)

type PubmedBackend struct {
}

type PubmedQuery struct {
	repr string
}

func (m PubmedQuery) Representation() (interface{}, error) {
	return m.repr, nil
}

func (m PubmedQuery) String() (string, error) {
	return m.repr, nil
}

func (m PubmedQuery) StringPretty() (string, error) {
	return m.repr, nil
}

func compilePubmed(q ir.BooleanQuery, level int) (l int, query PubmedQuery) {
	if q.Keywords == nil && len(q.Operator) == 0 {
		repr := ""
		for _, child := range q.Children {
			var comp PubmedQuery
			level, comp = compilePubmed(child, level)
			repr += comp.repr
		}
		return level, PubmedQuery{repr: repr}
	}

	childs := make([]string, len(q.Children))
	for i, child := range q.Children {
		l, comp := compilePubmed(child, level)
		level = l
		childs[i] = comp.repr
	}
	kwds := make([]string, len(q.Keywords))
	for i, keyword := range q.Keywords {
		var mf string
		qs := keyword.QueryString
		if len(keyword.Fields) == 1 && keyword.Fields[0] == "mesh_headings" {
			mf = "Mesh"
			if !keyword.Exploded {
				mf += ":noexp"
			}
		} else {
			for f, fields := range parser.PubMedFieldMapping {
				if len(fields) != len(keyword.Fields) {
					continue
				}
				match := true
				for i, field := range keyword.Fields {
					if field != fields[i] {
						match = false
					}
				}
				if match {
					mf = f
					break
				}
			}
			if len(mf) == 0 {
				mf = "All Fields"
			}
		}
		qs = fmt.Sprintf("%v[%v]", qs, mf)
		kwds[i] = qs
		level += 1
	}

	kwds = append(kwds, childs...)

	repr := fmt.Sprintf("(%v)", strings.Join(kwds, strings.ToUpper(fmt.Sprintf(" %v ", q.Operator))))
	level += 1
	return level, PubmedQuery{repr: repr}
}

func (b PubmedBackend) Compile(ir ir.BooleanQuery) (BooleanQuery, error) {
	_, q := compilePubmed(ir, 1)
	return q, nil
}

func NewPubmedBackend() PubmedBackend {
	return PubmedBackend{}
}
