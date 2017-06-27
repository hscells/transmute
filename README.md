# transmute

[![GoDoc](https://godoc.org/github.com/hscells/transmute?status.svg)](https://godoc.org/github.com/hscells/transmute)

_PubMed/Medline Query Transpiler_

The goal of transmute is to provide a way of normalising PubMed/Medline search strategies from systematic reviews. As a
result of the normalise process (into a _standard immediate representation_) is that the resulting normalised query can 
be analysed and transformed. This is why transmute is described as a _transpiler_. An immediate representation allows
trivial transformation to boolean queries acceptable by search engines, such as the Elasticsearch DSL, which is what the 
immediate representation is based off.

The parser of transmute attempts to automatically transform based on some heuristics, since search strategies can be 
reported in a variety of ways. Consider the following search strategies, transmute aims to parse them with as little 
human input as possible:
 
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

Both are valid Pubmed and Medline search strategies reported in real systematic reviews; transmute can currently parse
both into the same representation with close to no human input.

## Installing

As well as being a well-documented library, transmute can also be used on the command line. Since it is still in
development, it can be built from source with go tools:

```bash
go get -u github.com/hscells/transmute
cd $GOPATH/src/github.com/hscells/transmute
go build
./transmute --help
```

## Assumptions

The goal of transmute is to parse and transform PubMed/Medline queries into queries suitable for other search engines.
However, the parser makes some assumptions about the query:

 - PubMed/Medline field names (`.tw`, `[Mesh]`, etc.) are mapped to generic field names such as `title` and 
 `mesh_headings`.
 - If a field name cannot be mapped, you will receive a warning and the field will be added without being mapped. This
  allows queries such as `(a and b).abstract` to be transformed as (for example in ElasticSearch):
  ```json
  {
      "bool": {
          "disable_coord": true,
          "must": [
              {
                  "match": {
                      "abstract": "a"
                  }
              },
              {
                  "match": {
                      "abstract": "b"
                  }
              }
          ]
      }
  }
  ```
  - The parser does not attempt to simplify boolean expressions, so badly written queries will remain inefficient.
  
## Extending

If you would like to extend transmute and create a new backend for it, have a read of the 
[documentation](https://godoc.org/github.com/hscells/transmute/backend#Backend). As this should lead you in the right
direction. Writing a new backend requires the transformation of the immediate representation into the target query
language.