package parser

import (
	"os"
	"testing"
)

func TestReader(t *testing.T) {

	fs, err := os.Open("./../fixtures/config.dq")

	if err != nil {
		t.Error("File.Error: ", err)
	}

	ps := NewRunePacker(fs)

	g, err := ps.Read()

	if err != nil {
		t.Error("Parser.Read: ", err)
	}

	err = ps.Unread()

	if err != nil {
		t.Error("Parser.UnRead: ", err)
	}

	rd, err := ps.Read()

	if err != nil {
		t.Error("Parser.AfterUnRead.Read: ", err)
	}

	if g != rd {
		t.Fatalf("Unread failed with wrong packets: expecting %s found %s", g, rd)
	}

}
