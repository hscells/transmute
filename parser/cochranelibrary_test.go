package parser

import "testing"

func TestCochraneLibraryParse(t *testing.T) {
	cl := CochraneLibParser{}

	cl.TransformNested(`("lung cancer":tw)`, nil)
	cl.TransformNested(`(dialy?is:ti AND (kidney*:ti,ab NEAR/3 renal) AND "lung cancer"):tw,ab`, nil)
}
