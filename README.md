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

As well as being a well-documented library, transmute can also be used on the command line. Since it is still in
development, it can be built from source with go tools:

```bash
go get -u github.com/hscells/transmute
cd $GOPATH/src/github.com/hscells/transmute
go build
./transmute --help
```