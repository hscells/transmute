package parser

import (
	"testing"
	"fmt"
	"encoding/json"
)

func TestParse46(t *testing.T) {
	data := Load("../data/46")
	fmt.Println(data)
	query := Parse(data, ' ', '.')

	d, err := json.MarshalIndent(query, "", "\t")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(d))
}

//func TestParse288(t *testing.T) {
//	data := Load("../data/288")
//	fmt.Println(data)
//	query := Parse(data, ' ', '.')
//
//	d, err := json.MarshalIndent(query, "", "\t")
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println(string(d))
//}