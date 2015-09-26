package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"code.google.com/p/go-uuid/uuid"
)

//ErrBadQuery represents a badly split query
var ErrBadQuery = errors.New(BadQuery)

//OnlySpaces represents a regexp for only space values
var OnlySpaces = regexp.MustCompile(`^s+$`)
var onlyesc = regexp.MustCompile(`W+`)

func extractID(q string, keyword string) string {
	st, err := stripQuery(q)

	if err != nil || len(st) <= 0 {
		return uuid.New()
	}

	for _, v := range st {
		kve := strings.Split(v, ":")

		if len(kve) <= 0 {
			continue
		}

		ke, ve := strings.TrimSpace(kve[0]), strings.TrimSpace(kve[1])

		if ke != keyword {
			continue
		}

		return ve
	}

	return uuid.New()
}

func cleanQueryId(q string, keyword string) string {
	st, err := stripQuery(q)

	if err != nil || len(st) <= 0 {
		return q
	}

	for k, v := range st {
		kve := strings.Split(v, ":")

		if len(kve) <= 0 {
			continue
		}

		ke := strings.TrimSpace(kve[0])

		if ke != keyword {
			continue
		}

		st = append(st[0:k], st[k+1:]...)
		return fmt.Sprintf("(%s)", strings.Join(st, ","))
	}

	return q
}

func isUnique(q string) bool {
	st, err := stripQuery(q)

	if err != nil {
		return false
	}

	if len(st) <= 0 {
		return false
	}

	return true
}

func stripQuery(q string) ([]string, error) {
	var sl []string
	nk := strings.TrimSuffix(strings.TrimPrefix(q, "("), ")")

	sl = strings.Split(nk, ",")

	if len(sl) == 1 && sl[0] == "" {
		sl = sl[:0]
		return sl, nil
	}

	return sl, nil
}

func isSpace(c string) bool {
	return c == " " || c == "\t" || c == "\n"
}

func isLineBreak(c rune) bool {
	return c == '\r' || c == '\n'
}

func isWhiteSpace(c rune) bool {
	// log.Printf("checking ws: {%s}", c)
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isComma(c rune) bool {
	return c == ','
}

func isLetter(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isAlpha(c rune) bool {
	return isLetter(c) || isDigit(c) || isSymbol(c)
}

func isSymbol(c rune) bool {
	return c == '_'
}

func isDigit(c rune) bool {
	return (c >= '0' || c <= '9')
}

func isGroup(c rune) bool {
	return isGroupEnd(c) || isGroupStart(c)
}

func isGroupStart(c rune) bool {
	return c == '{'
}

func isGroupEnd(c rune) bool {
	return c == '}'
}

func isQuery(c rune) bool {
	return isQueryStart(c) || isQueryEnd(c)
}

func isQueryStart(c rune) bool {
	return c == '('
}

func isQueryEnd(c rune) bool {
	return c == ')'
}

func isSpecial(c rune) bool {
	return isQuery(c) || isGroup(c)
}

func report(msg, val string, line, pos int) error {
	return (fmt.Errorf(`
A disturbance in the Force in Line: %d, Pos: %d:
  -  Cause: '%s'
  -  Why: %s
`, line, pos, val, msg))
}
