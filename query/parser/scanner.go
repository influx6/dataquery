package parser

import (
	"bytes"
	"io"
)

type (
	//TokenType returns a token type
	TokenType int

	//Scanner provides the lexical phase
	Scanner struct {
		rd   *RunePacker
		line int
		pos  int
		// readlen int
		lockpos bool
		reads   []int
	}

	//Token provides a token type
	Token struct {
		Type   TokenType
		Data   string
		Pos    int
		Line   int
		Length int
	}
)

var (
	eof = rune(0)
)

const (
	//Invalid represent invalid types
	Invalid TokenType = iota
	//EOF represents eof
	EOF
	//WS represent whitespace
	WS

	//Comma represents comma ','
	Comma //,

	//Indent represent keywords
	Indent
	//Query representes basic (id:),(lt:,gt:) stacks
	Query
	//GroupStart represents a standard attribute {
	GroupStart
	//GroupEnd represents a standard attribute }
	GroupEnd
)

//NewToken returns a token and its type
func NewToken(c string, s TokenType, p, l int) *Token {
	return &Token{Data: c, Type: s, Pos: p, Line: l, Length: len(c)}
}

//EqualsType checks equality of token type
func (t *Token) EqualsType(b interface{}) bool {
	tok, ok := b.(*Token)

	if ok {
		return tok.Type == t.Type
	}

	to, ok := b.(TokenType)

	if ok {
		return t.Type == to
	}

	return false
}

//Equals checks equality of token type
func (t *Token) Equals(b interface{}) bool {
	tok, ok := b.(*Token)

	if ok {
		return tok.Type == t.Type && tok.Data == t.Data
	}

	return t.Data == b
}

//NewScanner returns a new scanner
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{rd: NewRunePacker(r)}
}

//Reset resets a scanner back to top
func (s *Scanner) Reset() {
	s.pos, s.line = -1, -1
	s.reads = s.reads[:0]
	s.rd.Reset()
	// s.rd.Reset(s.rx)
}

//read returns a rune and error
func (s *Scanner) read() (rune, error) {
	c, err := s.rd.Read()
	if err == nil {
		if s.lockpos {
			s.lockpos = false
			// s.pos++
		} else {
			if s.line == 0 {
				s.line++
			}
			s.pos++
		}
	}

	return c, err
}

//readOnly returns a rune
func (s *Scanner) readOnly() rune {
	r, _ := s.read()
	return r
}

func (s *Scanner) recordRead(tok *Token) *Token {
	s.reads = append(s.reads, len(tok.Data))
	return tok
}

func (s *Scanner) unreadLast() error {
	if len(s.reads) <= 0 {
		return nil
	}

	lastread := s.reads[len(s.reads)-1]
	if lastread <= 0 {
		return nil
	}

	var err error

	BackwardsIf(lastread, func(n int, stop func()) {
		s.freeze()
		err = s.unread()
		if err != nil {
			stop()
		}
	})

	s.reads = s.reads[0:(len(s.reads) - 1)]
	return err
}

//freeze freezes the line and pos values
func (s *Scanner) freeze() {
	s.lockpos = true
}

//unfreeze unfreezes the line and pos values
func (s *Scanner) unfreeze() {
	s.lockpos = false
}

//unread  the rune
func (s *Scanner) unread() error {
	if !s.lockpos {
		s.pos--
	}
	return s.rd.Unread()
}

//Scan is the public function for scanner that retuns each token and its type
func (s *Scanner) Scan() *Token {
	c, err := s.read()

	if err != nil {
		return NewToken(string(eof), EOF, s.pos, s.line)
		// return nil
	}

	if isWhiteSpace(c) {
		s.unread()
		return s.recordRead(s.scanWhiteSpace())
	} else if isAlpha(c) && !isSpecial(c) && !isComma(c) {
		s.unread()
		return s.recordRead(s.scanIdent())
	} else if isQueryStart(c) {
		s.unread()
		return s.recordRead(s.scanQuery())
	}

	switch c {
	case '{':
		return s.recordRead(NewToken(string(c), GroupStart, s.pos, s.line))
	case '}':
		return s.recordRead(NewToken(string(c), GroupEnd, s.pos, s.line))
	case ',':
		return s.recordRead(NewToken(string(c), Comma, s.pos, s.line))
	}

	return NewToken(string(eof), EOF, s.pos, s.line)
}

//scanWhiteSpace scans out the whitespace
func (s *Scanner) scanWhiteSpace() *Token {
	var buff bytes.Buffer

	for {

		if ch := s.readOnly(); ch == eof {
			return NewToken(string(eof), EOF, s.pos, s.line)
		} else if !isWhiteSpace(ch) {
			s.unread()
			break
		} else {
			if isLineBreak(ch) {
				s.line++
			}
			buff.WriteRune(ch)
		}
	}

	return NewToken(buff.String(), WS, s.pos, s.line)
}

//scanIdent scans out the group photo,name,id
func (s *Scanner) scanIdent() *Token {
	var buff bytes.Buffer

	for {

		if ch := s.readOnly(); ch == eof {
			// break
			return NewToken(string(eof), EOF, s.pos, s.line)
			// return nil
		} else if isSpecial(ch) || isWhiteSpace(ch) || isComma(ch) {
			s.unread()
			break
		} else if !isAlpha(ch) {
			s.unread()
			break
		} else {
			buff.WriteRune(ch)
		}

	}

	return NewToken(buff.String(), Indent, s.pos, s.line)
}

//scanQuery scans out the query (id:),(lt:,gt:)
func (s *Scanner) scanQuery() *Token {
	var buff bytes.Buffer

	for {

		if ch := s.readOnly(); ch == eof {
			// break
			return NewToken(string(eof), EOF, s.pos, s.line)
			// return nil
		} else if isQueryEnd(ch) {
			buff.WriteRune(ch)
			break
		} else {
			buff.WriteRune(ch)
		}

	}

	return NewToken(buff.String(), Query, s.pos, s.line)
}

//BackwardsIf takes a value and walks Backward till 0 unless the stop function is called
func BackwardsIf(to int, fx func(int, func())) {
	state := true
	for i := to; i > 0; i-- {
		if !state {
			break
		}
		fx(i, func() { state = false })
	}
}
