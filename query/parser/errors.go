package parser

import "fmt"

const (
	InvalidIndentStart  = "Invalid Start Line. Expected identifier type eg User(...)"
	InvalidIndentFollow = "Invalid token after Identifier expected a query '(..)' or body begin token '{' "

	InvalidStart      = "Invalid Start Character. Expected ('{')"
	InvalidEnd        = "Invalid End Character. Expected ('}')"
	InvalidQueryStart = "Invalid Character. Expected '('"
	InvalidQueryEnd   = "Invalid Character. Expected ')'"
	NoComma           = "Invalid Character. Expected (',')"
	InvalidComma      = "Invalid Character. UnExpected (',')"
	BadQuery          = "Invalid Formatted Query, pattern (id:40,...)"
	BadQuerySection   = "Invalid Formatted Query Option. Pattern should be 'id:400' i.e 'key:value', etc"
	EOFCase           = "Invalid Character. UnExpected EOF"
)

type (

	//SocketNotMadeError represents two nodes unable to be bounded
	SocketNotMadeError struct {
		to, from string
	}
)

//NewSocketNotMade returns a new sock error
func NewSocketNotMade(to, from string) SocketNotMadeError {
	return SocketNotMadeError{
		to:   to,
		from: from,
	}
}

//Error returns a string representation of the error
func (s SocketNotMadeError) Error() string {
	return fmt.Sprintf("BindError between %+s and %+s", s.to, s.from)
}
