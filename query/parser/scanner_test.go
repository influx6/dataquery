package parser

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func makeScanner(file string) *Scanner {
	bg, err := os.Open(file)

	if err != nil {
		log.Fatal("Error with scanner", err)
	}

	return NewScanner(bg)
}

var scanner = makeScanner("./../fixtures/config.dq")

func TestScannerRead(t *testing.T) {
	r, err := scanner.read()
	defer scanner.unread()

	if err != nil {
		t.Fatal(err)
	}

	if r != '{' {
		t.Fatalf("Invalid first rune %q", r)
	}
}

func TestScanGroupStart(t *testing.T) {
	token := scanner.Scan()

	if token.Data != "{" {
		t.Fatalf("Invalid recieved token %s", token.Data)
	}

}

func TestScanner(t *testing.T) {
	bg, err := os.Open("./../fixtures/models.dq")

	if err != nil {
		t.Fatal(err)
	}

	var scanner = NewScanner(bg)
	var buff bytes.Buffer
	count := 0

	for {
		tk := scanner.Scan()

		// log.Println("Token:", tk)
		if count <= 0 {
			scanner.unreadLast()
			count++
		}

		if tk.EqualsType(EOF) {
			break
		}

		buff.WriteString(tk.Data)
	}

	if scanner.pos != 344 {
		t.Fatalf("Wrong last position in file expected 219 got %d", scanner.pos)
	}

}
