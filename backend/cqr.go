package backend

import (
	"encoding/json"
	"github.com/hscells/cqr"
	"github.com/hscells/transmute/ir"
)

// CommonQueryRepresentationQuery is the transmute wrapper for CQR.
type CommonQueryRepresentationQuery struct {
	repr cqr.CommonQueryRepresentation
}

// CommonQueryRepresentationBackend is the backend for compiling transmute ir into CQR.
type CommonQueryRepresentationBackend struct{}

// Representation returns the CQR.
func (q CommonQueryRepresentationQuery) Representation() (interface{}, error) {
	return q.repr, nil
}

// String returns a JSON-encoded representation of the cqr.
func (q CommonQueryRepresentationQuery) String() (string, error) {
	b, err := json.Marshal(q.repr)
	return string(b), err
}

// StringPretty returns a pretty-printed JSON-encoded representation of the cqr.
func (q CommonQueryRepresentationQuery) StringPretty() (string, error) {
	b, err := json.MarshalIndent(q.repr, "", "    ")
	return string(b), err
}

// Compile transforms the transmute ir into CQR. The CQR is slightly different to the transmute ir, in that the
// depth of the children is different. Take note of how the children of a transmute ir differs from the children of CQR.
func (b CommonQueryRepresentationBackend) Compile(q ir.BooleanQuery) (BooleanQuery, error) {
	var children []cqr.CommonQueryRepresentation
	for _, keyword := range q.Keywords {
		k := cqr.NewKeyword(keyword.QueryString, keyword.Fields...).
			SetOption("exploded", keyword.Exploded).
			SetOption("truncated", keyword.Truncated)
		children = append(children, k)
	}
	for _, child := range q.Children {
		var subChildren []cqr.CommonQueryRepresentation
		for _, subChild := range child.Children {
			c, err := b.Compile(subChild)
			if err != nil {
				return nil, err
			}
			cqrSub := c.(CommonQueryRepresentationQuery).repr
			subChildren = append(subChildren, cqrSub)
		}
		for _, keyword := range child.Keywords {
			k := cqr.NewKeyword(keyword.QueryString, keyword.Fields...).
				SetOption("exploded", keyword.Exploded).
				SetOption("truncated", keyword.Truncated)
			subChildren = append(subChildren, k)
		}

		if len(child.Operator) == 0 {
			children = append(children, subChildren...)
		} else {
			bq := cqr.NewBooleanQuery(child.Operator, subChildren)
			for k, v := range child.Options {
				bq.SetOption(k, v)
			}
			children = append(children, bq)
		}
	}

	var repr cqr.CommonQueryRepresentation
	if len(q.Operator) == 0 && len(q.Children) == 1 {
		var keywords []cqr.CommonQueryRepresentation
		for _, kw := range q.Children[0].Keywords {
			keywords = append(keywords, cqr.NewKeyword(kw.QueryString, kw.Fields...).SetOption("exploded", kw.Exploded).SetOption("truncated", kw.Truncated))
		}

		for _, child := range q.Children[0].Children {
			keyword, err := b.Compile(child)
			if err != nil {
				return nil, err
			}
			keywords = append(keywords, keyword.(CommonQueryRepresentationQuery).repr)
		}
		repr = cqr.NewBooleanQuery(q.Children[0].Operator, keywords)
	} else {
		repr = cqr.NewBooleanQuery(q.Operator, children)
	}

	for k, v := range q.Options {
		repr.SetOption(k, v)
	}

	return CommonQueryRepresentationQuery{repr: repr}, nil
}

// NewCQRBackend returns a new CQR backend.
func NewCQRBackend() CommonQueryRepresentationBackend {
	return CommonQueryRepresentationBackend{}
}

func NewCQRQuery(query cqr.CommonQueryRepresentation) CommonQueryRepresentationQuery {
	return CommonQueryRepresentationQuery{repr: query}
}
