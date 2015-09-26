package parser

import (
	"os"
	"testing"

	"github.com/influx6/ds"
	"github.com/influx6/flux"
)

func TestChunkScanning(t *testing.T) {
	fs, err := os.Open("./../fixtures/config.dq")

	if err != nil {
		flux.FatalFailed(t, "File.Error occured: %+s", err)
	}

	err = ScanChunks(NewScanner(fs), func(cu string) {
		// t.Logf("Chunks: %s", cu)
	})

	if err != nil {
		flux.FatalFailed(t, "ChunkScanning.Error occured: %+s", err)
	}

	flux.LogPassed(t, "ChunkScanning finished successfully!")
}

func TestBadFileParser(t *testing.T) {
	ps := NewParser(DefaultInspectionFactory)

	fs, err := os.Open("./../fixtures/config.dq")

	if err != nil {
		flux.FatalFailed(t, "File.Error occured: %+s", err)
	}

	_, err = ps.Scan(fs)

	if err == nil {
		flux.FatalFailed(t, "Parser is supposed to error out about bad file")
	}

	flux.LogPassed(t, "Bad file failed properly: %+s", err)
}

func TestModels(t *testing.T) {
	ps := NewParser(DefaultInspectionFactory)

	fs, err := os.Open("./../fixtures/models.dq")

	if err != nil {
		flux.FatalFailed(t, "File.Error occured: %+s", err)
	}

	g, err := ps.Scan(fs)

	if err != nil {
		flux.FatalFailed(t, "Parser is supposed to not error out about model file: %+s", err)
	}

	flux.LogPassed(t, "Generated query graph successfully: %t", g != nil)
}

func TestGoodFileParser(t *testing.T) {
	ps := NewParser(DefaultInspectionFactory)

	fs, err := os.Open("./../fixtures/dataset.dq")

	if err != nil {
		flux.FatalFailed(t, "File.Error occured: %+s", err)
	}

	g, err := ps.Scan(fs)

	if err != nil {
		flux.FatalFailed(t, "Parser.Error: ", err)
	}

	flux.LogPassed(t, "Generated query graph successfully: %t", g != nil)

	itr, err := ds.CreateGraphTransversor(ds.DFPreOrderDirective(nil, nil))

	if err != nil {
		flux.FatalFailed(t, "Transverso.Error occured: %+s", err)
	}

	itr.Use(g.Get("user"))

	for itr.Next() == nil {
		node := itr.Node().(*ParseNode)
		t.Logf("Node-Attr: %s %+s", node.Name(), node.Attr)
		t.Logf("Node-Record: %s %+s", node.Name(), node.Records.Keys())
		t.Logf("Node-Rules: %s %+s", node.Name(), node.Rules)
	}

	flux.LogPassed(t, "Graph Iterator works properly")
}

func TestModelFileParser(t *testing.T) {
	ps := NewParser(DefaultInspectionFactory)

	fs, err := os.Open("./../fixtures/models.dq")

	if err != nil {
		flux.FatalFailed(t, "File.Error occured: %+s", err)
	}

	g, err := ps.Scan(fs)

	if err != nil {
		flux.FatalFailed(t, "Parser.Error occured: %+s", err)
	}

	itr, err := ds.CreateGraphTransversor(ds.DFPreOrderDirective(nil, nil))

	if err != nil {
		t.Fatal(err)
	}

	itr.Use(g.Get("user"))

	for itr.Next() == nil {
		node := itr.Node().(*ParseNode)
		t.Logf("Node-Attr: %s %+s", node.Name(), node.Attr)
		t.Logf("Node-Record: %s %+s", node.Name(), node.Records.Keys())
		t.Logf("Node-Rules: %s %+s", node.Name(), node.Rules)
	}

	flux.LogPassed(t, "Successfully passed model query file properly")
}
