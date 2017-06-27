// Package parser implements a parser for the search strategies in systematic reviews. The goal of the parser is to
// transform it into an immediate representation that can then be translated into queries suitable for other systems.
//
// This file contains utilities relating to the parsing.
package parser

import (
	"unicode"
	"log"
	"fmt"
)

var (
	fieldMap = map[string][]string{
		"mp": []string{"title", "abstract", "mesh_headings"},
		"af": []string{"title", "abstract", "mesh_headings"},
		"tw": []string{"title", "abstract"},
		"nm": []string{"abstract", "mesh_headings"},
		"ab": []string{"abstract"},
		"ti": []string{"title"},
		"ot": []string{"title"},
		"sh": []string{"mesh_headings"},
		"px": []string{"mesh_headings"},
		"rs": []string{"mesh_headings"},
		"fs": []string{"mesh_headings"},
		"rn": []string{"mesh_headings"},
		"kf": []string{"mesh_headings"},
		"sb": []string{"mesh_headings"},
		"mh": []string{"mesh_headings"},
		"pt": []string{"pubtype"},

		"tiab": []string{"title", "abstract"},

		"Mesh": []string{"mesh_headings"},
		"Title": []string{"title"},
		"Title/Abstract": []string{"title", "abstract"},
	}
)

func LookupField(f string) []string {
	if fields, ok := fieldMap[f]; !ok {
		log.Println(fmt.Sprintf("WARN: %v is not a valid field.", f))
		return []string{f}
	} else {
		return fields
	}
}

// IsUpper tests to see if a string consists of all uppercase characters.
func IsUpper(s string) bool {
	for _, c := range s {
		if unicode.IsLower(c) {
			return false
		}
	}
	return true
}

// IsLower tests to see if a string consists of all lowercase characters.
func IsLower(s string) bool {
	for _, c := range s {
		if unicode.IsUpper(c) {
			return false
		}
	}
	return true
}

// IsOperator tests to see if a string is a valid PubMed/Medline operator.
func IsOperator(s string) bool {
	return s == "or" ||
		s == "and" ||
		s == "not" ||
		s == "adj" ||
		s == "adj2" ||
		s == "adj3" ||
		s == "adj4" ||
		s == "adj5" ||
		s == "adj6" ||
		s == "adj7" ||
		s == "adj8"
}