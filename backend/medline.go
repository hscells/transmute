package backend

import (
	"github.com/hscells/transmute/ir"
	"fmt"
	"strings"
	"strconv"
	"github.com/hscells/transmute/parser"
)

type MedlineBackend struct {
}

type MedlineQuery struct {
	repr string
}

func (m MedlineQuery) Representation() (interface{}, error) {
	return m.repr, nil
}

func (m MedlineQuery) String() (string, error) {
	return m.repr, nil
}

func (m MedlineQuery) StringPretty() (string, error) {
	return m.repr, nil
}

func compileMedline(q ir.BooleanQuery, level int) (l int, query MedlineQuery) {
	repr := ""
	var op []string
	if q.Keywords == nil && len(q.Operator) == 0 {
		for _, child := range q.Children {
			var comp MedlineQuery
			level, comp = compileMedline(child, level)
			repr += comp.repr
		}
		return level, MedlineQuery{repr: repr}
	}
	for _, child := range q.Children {
		l, comp := compileMedline(child, level)
		repr += comp.repr
		level = l
		op = append(op, strconv.Itoa(l-1))
	}
	for _, keyword := range q.Keywords {
		var mf string
		qs := keyword.QueryString
		if keyword.Exploded {
			qs = "exp " + qs
		}
		if len(keyword.Fields) == 1 && keyword.Fields[0] == "mesh_headings" {
			qs += "/"
		} else {
			for f, fields := range parser.MedlineFieldMapping {
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
			qs = fmt.Sprintf("%v.%v.", qs, mf)
		}
		qs = strings.Replace(qs, "~", "$", -1)
		repr += fmt.Sprintf("%v. %v\n", level, qs)
		op = append(op, strconv.Itoa(level))
		level += 1
	}
	if len(op) > 0 {
		repr += fmt.Sprintf("%v. %v\n", level, strings.Join(op, fmt.Sprintf(" %v ", q.Operator)))
	}
	level += 1
	return level, MedlineQuery{repr: repr}
}

func (b MedlineBackend) Compile(ir ir.BooleanQuery) (BooleanQuery, error) {
	_, q := compileMedline(ir, 1)
	return q, nil
}

func NewMedlineBackend() MedlineBackend {
	return MedlineBackend{}
}
