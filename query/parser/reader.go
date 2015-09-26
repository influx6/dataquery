package parser

import (
	"bufio"
	"errors"
	"io"
)

type (

	//RunePacker provides a means of reading and unreading runes efficienctly
	RunePacker struct {
		reader  io.Reader
		rd      *bufio.Reader
		buffer  []rune
		unreads int
	}
)

var (
	maxBuffer = 4096
	//ErrNoUnread refers to a packer unable to unread anymore
	ErrNoUnread = errors.New("Unable to unread passed point!")
)

//NewRunePacker returns a new runepacker
func NewRunePacker(r io.Reader) *RunePacker {
	return &RunePacker{
		reader:  r,
		rd:      bufio.NewReader(r),
		buffer:  make([]rune, 0),
		unreads: 0,
	}
}

//records adds a rune into the buffer
func (r *RunePacker) record(s rune) {
	if len(r.buffer) >= maxBuffer {
		r.buffer = r.buffer[1:]
	}
	r.buffer = append(r.buffer, s)
}

//Close resets the reader
func (r *RunePacker) Close() error {
	return r.Reset()
}

//Reset resets the reader and internal buffer
func (r *RunePacker) Reset() error {
	r.buffer = r.buffer[:0]
	r.rd.Reset(r.reader)
	return nil
}

//Unread backtracks the reader to previously scanned runes
func (r *RunePacker) Unread() error {
	if r.unreads < len(r.buffer) {
		r.unreads++
		return nil
	}
	return ErrNoUnread
}

//Read reads the next rune
func (r *RunePacker) Read() (rune, error) {
	if r.unreads > 0 {
		d := r.buffer[len(r.buffer)-r.unreads]
		r.unreads--
		return d, nil
	}

	d, _, err := r.rd.ReadRune()

	if err == nil {
		r.record(d)
	}

	return d, err
}
