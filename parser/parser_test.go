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

	d, err := json.MarshalIndent(query, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(d))
}

func TestParse450(t *testing.T) {
	data := Load("../data/450")
	fmt.Println(data)
	query := Parse(data, rune(0), rune(0))

	d, err := json.MarshalIndent(query, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(d))
}

func TestParse433(t *testing.T) {
	data := Load("../data/433")
	fmt.Println(data)
	query := Parse(data, rune(0), rune(0))

	d, err := json.MarshalIndent(query, "", "    ")
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