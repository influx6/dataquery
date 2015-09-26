package adaptors

import (
	"sync"
	"testing"

	"github.com/influx6/dataquery/parser"
	"github.com/influx6/ds"
	"github.com/influx6/flux"
)

func TestFileParser(t *testing.T) {
	var ws sync.WaitGroup
	ws.Add(1)

	fs := FileParseAdaptor(parser.DefaultInspectionFactory)

	fs.React(func(rs flux.Reactor, err error, data interface{}) {
		ws.Done()
		if err != nil {
			flux.FatalFailed(t, "Error occured string: %v", err.Error())
		}

		if _, ok := data.(ds.Graphs); !ok {
			flux.FatalFailed(t, "Expected type ds.Graphs: %v", data)
		}

		flux.LogPassed(t, "Completed with parser Graph")
	}, true)

	fs.Send("./../fixtures/dataset.dq")

	ws.Wait()
	fs.Close()
}

func TestChunkScannerWithSinglePack(t *testing.T) {
	var ws sync.WaitGroup
	ws.Add(1)

	fs := ChunkFileScanAdaptor()

	fs.React(func(rs flux.Reactor, err error, data interface{}) {
		ws.Done()
		if err != nil {
			flux.FatalFailed(t, "Error occured string: %v", err.Error())
		}

		if _, ok := data.(string); !ok {
			t.Fatalf("Expected type string: %v", data)
		}

		flux.LogPassed(t, "Completed with scanning on SinglePack file")
	}, true)

	fs.Send("./../fixtures/dataset.dq")

	ws.Wait()

	fs.Close()
}

func TestChunkParserCombo(t *testing.T) {
	var ws sync.WaitGroup
	ws.Add(1)

	fs := ChunkFileScanAdaptor()
	ps := ParseAdaptor(parser.DefaultInspectionFactory)

	fs.Bind(ps, true)

	fs.React(func(rs flux.Reactor, err error, data interface{}) {
		ws.Done()

		if err != nil {
			flux.FatalFailed(t, "Error occured string: %v", err.Error())
		}

		if _, ok := data.(ds.Graphs); ok {
			flux.FatalFailed(t, "Expected type ds.Graphs: %v", data)
		}

		flux.LogPassed(t, "Completed with parser Graph from ChunckFileScanner and Parser Combo")
	}, true)

	fs.Send("./../fixtures/dataset.dq")

	ws.Wait()
	fs.Close()
}

func TestChunkScannerWithLargePack(t *testing.T) {
	var ws sync.WaitGroup
	ws.Add(2)

	fs := ChunkScanAdaptor()

	fs.React(func(rs flux.Reactor, err error, data interface{}) {
		ws.Done()

		if err != nil {
			flux.FatalFailed(t, "Error occured string: %v", err.Error())
		}

		if _, ok := data.(string); !ok {
			t.Fatalf("Expected type string: %v", data)
		}

		flux.LogPassed(t, "Completed with scanning with ChunkScanner")
	}, true)

	fs.Send(`
			{
			  user(){
			    id(is: 4000),
			    name,
			    state,
			    address,
			    skills(range: 30..100),
			    age(lt:30, gte:40),
			    age(is: 20),
			    day(isnot: wednesday),
			    photos(width: 400){
			      day,
			      fax,
			    },
			  },
			  admin(id:4,rack:10){
			    name,
			    email,
			    group,
			    levels,
			    permissions(){
			      code,
			      active,
			    },
			  },
			}
	`)

	ws.Wait()

	fs.Close()
}
