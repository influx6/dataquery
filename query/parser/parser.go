package parser

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/influx6/data/query/utils"
	ds "github.com/influx6/ds"
)

// NodeType represents a node type (target,query)
type NodeType int

const (
	//IDMarker is used to mark the attr to contain the user defined id of a node
	IDMarker = "id"
	//NONE represents an unknow type
	NONE NodeType = iota
	//MODELROOT represent the root node of all query nodes
	MODELROOT
	//MODELSUBROOT represent the sub root node of an embedded query in a root query
	MODELSUBROOT
)

// ParseNode defines a node used in the parser
type ParseNode struct {
	ds.Nodes
	name           string
	Key            string
	Parent         string
	PKey           string
	NType          NodeType
	Attr           *ds.StringSet
	Rules, Records *Collectors
	Result         []map[string]interface{}
}

//NewParseNode returns a new ParseNode instance
func NewParseNode(tp NodeType, val, pa, pk string, gs ds.Graphs) *ParseNode {
	alias := strings.ToLower(utils.RandomAlias())
	return &ParseNode{
		name:    val,
		Key:     alias,
		NType:   tp,
		Parent:  pa,
		PKey:    pk,
		Nodes:   ds.NewGraphNode(val, gs),
		Attr:    ds.NewStringSet(),
		Rules:   NewCollectors(),
		Records: NewCollectors(),
	}
}

//Name returns the tag/name of this node
func (p *ParseNode) Name() string {
	return p.name
}

//String returns a string representation of the node
func (p *ParseNode) String() string {
	smp := []string{
		p.Nodes.String(),
		fmt.Sprintf(", Records: %+s", p.Records),
		fmt.Sprintf(", Attr: %+s", p.Attr),
		fmt.Sprintf(", Rules: %+s", p.Rules),
	}

	return strings.Join(smp, "")
}

// Parser represents the baseline parser for query
type Parser struct {
	inspect *InspectionFactory
	skip    int
}

//NewParser returns a new instance parser
func NewParser(inspect *InspectionFactory) *Parser {
	return &Parser{
		inspect: inspect,
	}
}

//Scan scans the provided input and returns a graph
func (p *Parser) Scan(rl io.Reader) (ds.Graphs, error) {
	scan := NewScanner(rl)

	tok := scanOutWhiteSpace(scan)

	// log.Printf("indent-token", tok.Data, tok.Type)
	if !tok.EqualsType(Indent) {
		return nil, report(InvalidIndentStart, tok.Data, tok.Line, tok.Pos)
	}

	gos := ds.NewGraph()

	psn := NewParseNode(MODELROOT, tok.Data, "", "", gos)
	// gos.Add(tok.Data)
	gos.AddNode(psn)

	if err := scanSection(psn, gos, scan, p); err != nil {
		return nil, err
	}

	return gos, nil
}

func scanOutWhiteSpace(scan *Scanner) (tok *Token) {
	tok = scan.Scan()

	if tok.EqualsType(WS) {
		tok = scan.Scan()
	}
	return
}

func scanIdentWithQuery(data string, target *ParseNode, inspect *InspectionFactory) error {
	parts, err := stripQuery(data)

	if err != nil {
		return err
	}

	if len(parts) <= 0 {
		return nil
	}

	for _, v := range parts {
		rsv := strings.Split(v, ":")

		if len(rsv) < 2 {
			target.Attr.Add(v)
			continue
		}

		tag, value := strings.TrimSpace(rsv[0]), strings.Join(rsv[1:], "")

		tag = strings.ToLower(tag)
		// if !inspect.Has(tag) {
		// 	tag = "is"
		// }

		in, err := inspect.Find(tag)

		if err != nil {
			var cols []Collector
			cols = append(cols, Collector{
				"type":  "is",
				"value": strings.TrimSpace(value),
			})

			target.Rules.Set(tag, cols)
			continue
		}

		col, err := in.Create(value)

		if err != nil {
			return err
		}

		target.Rules.Set(tag, []Collector{col})
	}

	return nil
}

func scanAttrWithQuery(attr, query string, target *ParseNode, inspect *InspectionFactory) error {

	var conds []Collector

	parts, err := stripQuery(query)

	if err != nil {
		return err
	}

	if len(parts) <= 0 {
		return nil
	}

	for _, v := range parts {
		rsv := strings.Split(v, ":")

		if len(rsv) != 2 {
			return errors.New(BadQuerySection)
		}

		tag, value := strings.TrimSpace(rsv[0]), strings.Join(rsv[1:], "")
		// tag, value := strings.TrimSpace(rsv[0]), rsv[1]

		tag = strings.ToLower(tag)
		if !inspect.Has(tag) {
			tag = "is"
		}

		// //log.Debug("Checking for tag: %s", tag)
		in, err := inspect.Find(strings.ToLower(tag))
		// //log.Debug("Checking for tag: %s", tag, err)

		if err != nil {
			// continue
			return err
		}

		col, err := in.Create(value)

		if err != nil {
			return err
		}

		conds = append(conds, col)
	}

	target.Records.Set(attr, conds)

	return nil
}

func scanSection(target *ParseNode, graph ds.Graphs, scan *Scanner, p *Parser) error {

	tok := scanOutWhiteSpace(scan)
	// log.Printf("section-token", tok.Data, tok.Type)

	if !tok.EqualsType(Query) && !tok.EqualsType(GroupStart) {
		return report(InvalidIndentFollow, tok.Data, tok.Line, tok.Pos)
	}

	if tok.EqualsType(Query) {
		// log.Printf("Handler query for:", target.Name(), tok)
		scanIdentWithQuery(tok.Data, target, p.inspect)
		nxt := scanOutWhiteSpace(scan)

		if !nxt.EqualsType(GroupStart) {
			// log.Printf("did not see start{}:", target.Name(), tok)
			return report(InvalidStart, tok.Data, tok.Line, tok.Pos)
		}
	}

	for {

		curtok := scanOutWhiteSpace(scan)
		// log.Printf("scan-indent-token:", tok.Data)

		if curtok.EqualsType(EOF) {
			break
		}

		if curtok.EqualsType(Comma) {
			continue
		}

		if curtok.EqualsType(GroupEnd) {
			break
		}

		if curtok.EqualsType(Indent) {
			// log.Printf("found-indent-token:", curtok.Data)

			tag := curtok.Data
			// log.Printf("in-indent-level", curtok.Data)

			nx := scanOutWhiteSpace(scan)
			// log.Printf("in-indent-gs-data", nx.Data)

			if nx.EqualsType(Query) {
				nxx := scanOutWhiteSpace(scan)
				// log.Printf("in-indent-start", nxx.Data)

				if nxx.EqualsType(GroupStart) {
					scan.unreadLast()
					scan.unreadLast()

					psn := NewParseNode(MODELSUBROOT, curtok.Data, target.Name(), target.Key, graph)
					// graph.Add(tag)
					graph.AddNode(psn)

					// curnode := graph.Get(tag)
					// curnode := psn

					graph.BindNodes(target, psn, 0)

					err := scanSection(psn, graph, scan, p)

					if err != nil {
						return err
					}

					continue
				}

				scanAttrWithQuery(tag, nx.Data, target, p.inspect)
				continue
			}

			target.Records.Set(tag, nil)
		}

	}

	return nil
}

//ScanChunks takes a scanner and scans out each section of the supported query format
func ScanChunks(scan *Scanner, chunks func(string)) error {
	if chunks == nil {
		return nil
	}

	tok := scanOutWhiteSpace(scan)

	if tok.EqualsType(Invalid) || tok.EqualsType(EOF) {
		return ErrBadQuery
	}

	if tok.EqualsType(GroupStart) {
		return ScanChunks(scan, chunks)
	}

	scan.unreadLast()

	for {

		tok := scanOutWhiteSpace(scan)

		if tok.EqualsType(Invalid) {
			break
		}

		if tok.EqualsType(EOF) {
			break
		}

		if tok.EqualsType(GroupEnd) {
			break
		}

		if tok.EqualsType(Comma) {
			continue
		}

		scan.unreadLast()
		cu, err := scanChunk(scan)

		if err != nil {
			return err
		}

		chunks(cu)
	}

	return nil
}

//scanChunk scans chunks of complete querys from a large text file of them
func scanChunk(scan *Scanner) (string, error) {
	var chunk []string
	open := 0
	isopen := false

	for {

		tok := scan.Scan()

		if tok.EqualsType(EOF) {
			break
		}

		if tok.EqualsType(GroupStart) {
			open++
			isopen = true
		}

		if tok.EqualsType(GroupEnd) {
			open--
			if open <= 0 {
				break
			}

			if isopen {
				chunk = append(chunk, tok.Data)
				isopen = false
			}
		}

		chunk = append(chunk, tok.Data)
	}

	return strings.TrimSuffix(strings.TrimSpace(strings.Join(chunk, "")), ","), nil
}
