<img height="200px" src="gopher.png" alt="gopher" align="right"/>

# transmute

[![GoDoc](https://godoc.org/github.com/hscells/transmute?status.svg)](https://godoc.org/github.com/hscells/transmute)
[![Go Report Card](https://goreportcard.com/badge/github.com/hscells/transmute)](https://goreportcard.com/report/github.com/hscells/transmute)
[![gocover](http://gocover.io/_badge/github.com/hscells/transmute)](https://gocover.io/github.com/hscells/transmute)

_PubMed/Medline Query Transpiler_

The goal of transmute is to provide a way of transforming PubMed/Medline search strategies from systematic reviews into
other queries suitable for other search engines. The result of the transformation is an _immediate representation_ which
can be analysed with greater ease or transformed again run on other search engines. This is why transmute is described
as a _transpiler_. An immediate representation allows trivial transformation to boolean queries acceptable by search
engines, such as Elasticsearch.

An example of a Medline and Pubmed query are:
 
```
1. MMSE*.ti,ab.
2. sMMSE.ti,ab.
3. Folstein*.ti,ab.
4. MiniMental.ti,ab.
5. \"mini mental stat*\".ti,ab.
6. or/1-5
```

```
(\"Contraceptive Agents, Female\"[Mesh] OR \"Contraceptive Devices, Female\"[Mesh] OR contracept*[tiab]) AND (\"Body Weight\"[Mesh] OR weight[tiab] OR \"Body Mass Index\"[Mesh]) NOT (cancer*[ti] OR polycystic [ti] OR exercise [ti] OR physical activity[ti] OR postmenopaus*[ti])
```

Both are valid Pubmed and Medline search strategies reported in real systematic reviews; transmute can currently
transform both Medline and PubMed queries. An example API usage by constructing a pipeline and executing it is shown in
the next section.

## API Usage

Here we construct a pipeline in Go:

```go
query := `1. MMSE*.ti,ab.
2. sMMSE.ti,ab.
3. Folstein*.ti,ab.
4. MiniMental.ti,ab.
5. \"mini mental stat*\".ti,ab.
6. or/1-5`

p := transmute.pipeline.NewPipeline(transmute.parser.NewMedlineParser(),
                                    transmute.backend.NewElasticsearchCompiler(),
                                    transmute.pipeline.TransmutePipelineOptions{RequiresLexing: true})
dsl, err := p.Execute(query)
if err != nil {
    panic(err)
}

println(dsl.StringPretty())
```

Which results in:

```json
{
    "query": {
        "bool": {
            "disable_coord": true,
            "should": [
                {
                    "bool": {
                        "should": [
                            {
                                "wildcard": {
                                    "title": "MMSE*"
                                }
                            },
                            {
                                "wildcard": {
                                    "abstract": "MMSE*"
                                }
                            }
                        ]
                    }
                },
                {
                    "multi_match": {
                        "fields": [
                            "title",
                            "abstract"
                        ],
                        "query": "sMMSE"
                    }
                },
                {
                    "bool": {
                        "should": [
                            {
                                "wildcard": {
                                    "title": "Folstein*"
                                }
                            },
                            {
                                "wildcard": {
                                    "abstract": "Folstein*"
                                }
                            }
                        ]
                    }
                },
                {
                    "multi_match": {
                        "fields": [
                            "title",
                            "abstract"
                        ],
                        "query": "MiniMental"
                    }
                },
                {
                    "bool": {
                        "should": [
                            {
                                "wildcard": {
                                    "title": "\"mini mental stat*\""
                                }
                            },
                            {
                                "wildcard": {
                                    "abstract": "\"mini mental stat*\""
                                }
                            }
                        ]
                    }
                }
            ]
        }
    }
}
```

## Command Line Usage

As well as being a well-documented library, transmute can also be used on the command line. Since it is still in
development, it can be built from source with go tools:

```bash
go get -u github.com/hscells/transmute
cd $GOPATH/src/github.com/hscells/transmute
go build
./transmute --help
./transmute --input mmse.query --parser medline --backend elasticsearch
```

The output of the command line pretty-prints the same output from above.

## Assumptions

The goal of transmute is to parse and transform PubMed/Medline queries into queries suitable for other search engines.
However, the project makes some assumptions about the query:

  - The parser does not attempt to simplify boolean expressions, so badly written queries will remain inefficient.
  - A query cannot compile to Elasticsearch when it contains an adjacency operator with more than one field. This is
  due to a limitation with Elasticsearch.
  
## Extending

If you would like to extend transmute and create a new backend for it, have a read of the 
[documentation](https://godoc.org/github.com/hscells/transmute/backend#Backend). As this should lead you in the right
direction. Writing a new backend requires the transformation of the immediate representation into the target query
language.

## Logo

The Go gopher was created by [Renee French](https://reneefrench.blogspot.com/), licensed under
[Creative Commons 3.0 Attributions license](https://creativecommons.org/licenses/by/3.0/).