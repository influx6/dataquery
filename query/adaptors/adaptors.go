package adaptors

import (
	"bytes"
	"errors"
	"os"

	"github.com/influx6/dataquery/parser"
	"github.com/influx6/ds"
	"github.com/influx6/flux"
)

// ErrGraphType defines a graph error type
var ErrGraphType = errors.New("Interface{} is not a ds.Graphs type")

// ErrInputTytpe defines an error when the interface type is not a string
var ErrInputTytpe = errors.New("Wrong-Type: input not a string")

// ChunkFileParser returns the a combined Reactor that provides the functions of a chunkfile(file containing  multiple query requests) scanner and parser
func ChunkFileParser(ds *parser.InspectionFactory) flux.Reactor {
	ms := flux.ReactorStack()

	ms.Bind(ChunkFileScanAdaptor(), true)
	ms.Bind(ParseAdaptor(ds), true)
	return ms
}

// ChunkParser returns a Reactor that combines the ChunkScanAdaptor with the ParseAdaptor to allow stringed queries
func ChunkParser(ds *parser.InspectionFactory) flux.Reactor {
	ms := flux.ReactorStack()

	ms.Bind(ChunkScanAdaptor(), true)
	ms.Bind(ParseAdaptor(ds), true)
	return ms
}

// ChunkFileScanAdaptor provides a Stacks for parser.Parser
func ChunkFileScanAdaptor() flux.Reactor {
	return flux.Reactive(func(v flux.Reactor, err error, d interface{}) {
		if err != nil {
			v.ReplyError(err)
			return
		}
		var data string
		var ok bool

		if data, ok = d.(string); !ok {
			v.ReplyError(ErrInputTytpe)
			return
		}

		var fs *os.File

		if fs, err = os.Open(data); err != nil {
			v.ReplyError(err)
			return
		}

		if err = parser.ScanChunks(parser.NewScanner(fs), func(query string) {
			v.Reply(query)
		}); err != nil {
			v.ReplyError(err)
		}
	})
}

// ChunkScanAdaptor provides a Stacks for parser.Parser and scans strings inputs for query
func ChunkScanAdaptor() flux.Reactor {
	return flux.Reactive(func(v flux.Reactor, err error, d interface{}) {
		if err != nil {
			v.ReplyError(err)
			return
		}

		var data string
		var ok bool

		if data, ok = d.(string); !ok {
			v.ReplyError(ErrInputTytpe)
			return
		}

		scan := parser.NewScanner(bytes.NewBufferString(data))

		if err = parser.ScanChunks(scan, func(query string) {
			v.Reply(query)
		}); err != nil {
			v.ReplyError(err)
		}
	})
}

// Parser provides a parser adaptor
type Parser struct {
	flux.Reactor
	parser *parser.Parser
}

// FileParseAdaptor provides a Stacks for parser.Parser
func FileParseAdaptor(inspect *parser.InspectionFactory) *Parser {
	ps := parser.NewParser(inspect)

	ad := flux.Reactive(func(v flux.Reactor, err error, d interface{}) {
		if err != nil {
			v.ReplyError(err)
			return
		}

		var data string
		var ok bool

		if data, ok = d.(string); !ok {
			v.ReplyError(ErrInputTytpe)
			return
		}

		var fs *os.File

		if fs, err = os.Open(data); err != nil {
			v.ReplyError(err)
			return
		}

		var gs ds.Graphs

		if gs, err = ps.Scan(fs); err != nil {
			v.ReplyError(err)
			return
		}

		v.Reply(gs)
	})

	return &Parser{
		Reactor: ad,
		parser:  ps,
	}
}

// ParseAdaptor provides a Stacks for parser.Parser to parse stringed queries rather than from a file,it takes a string of a full single query and parses it
func ParseAdaptor(inspect *parser.InspectionFactory) *Parser {
	ps := parser.NewParser(inspect)

	ad := flux.Reactive(func(v flux.Reactor, err error, d interface{}) {
		if err != nil {
			v.ReplyError(err)
			return
		}

		var data string
		var ok bool

		if data, ok = d.(string); !ok {
			v.ReplyError(ErrInputTytpe)
			return
		}

		var gs ds.Graphs

		if gs, err = ps.Scan(bytes.NewBufferString(data)); err != nil {
			v.ReplyError(err)
			return
		}

		v.Reply(gs)
	})

	return &Parser{
		Reactor: ad,
		parser:  ps,
	}
}

// QueryHandler provides a function type handler
type QueryHandler func(flux.Reactor, ds.Graphs)

// QueryAdaptor provides a simple sql parser
func QueryAdaptor(gx QueryHandler) flux.Reactor {
	return flux.Reactive(func(v flux.Reactor, err error, d interface{}) {
		if err != nil {
			v.ReplyError(err)
			return
		}

		da, ok := d.(ds.Graphs)

		if !ok {
			v.ReplyError(ErrGraphType)
			return
		}

		gx(v, da)
	})
}
