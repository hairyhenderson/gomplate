package cli

import (
	"strings"

	"fmt"
)

type uTokenType string

const (
	utPos        uTokenType = "Pos"
	utOpenPar    uTokenType = "OpenPar"
	utClosePar   uTokenType = "ClosePar"
	utOpenSq     uTokenType = "OpenSq"
	utCloseSq    uTokenType = "CloseSq"
	utChoice     uTokenType = "Choice"
	utOptions    uTokenType = "Options"
	utRep        uTokenType = "Rep"
	utShortOpt   uTokenType = "ShortOpt"
	utLongOpt    uTokenType = "LongOpt"
	utOptSeq     uTokenType = "OptSeq"
	utOptValue   uTokenType = "OptValue"
	utDoubleDash uTokenType = "DblDash"
)

type uToken struct {
	typ uTokenType
	val string
	pos int
}

func (t *uToken) String() string {
	return fmt.Sprintf("%s('%s')@%d", t.typ, t.val, t.pos)
}

type parseError struct {
	input string
	msg   string
	pos   int
}

func (t *parseError) ident() string {
	return strings.Map(func(c rune) rune {
		switch c {
		case '\t':
			return c
		default:
			return ' '
		}
	}, t.input[:t.pos])
}
func (t *parseError) Error() string {
	return fmt.Sprintf("Parse error at position %d:\n%s\n%s^ %s",
		t.pos, t.input, t.ident(), t.msg)
}

func uTokenize(usage string) ([]*uToken, *parseError) {
	pos := 0
	res := []*uToken{}
	var (
		tk = func(t uTokenType, v string) {
			res = append(res, &uToken{t, v, pos})
		}

		tkp = func(t uTokenType, v string, p int) {
			res = append(res, &uToken{t, v, p})
		}

		err = func(msg string) *parseError {
			return &parseError{usage, msg, pos}
		}
	)
	eof := len(usage)
	for pos < eof {
		switch c := usage[pos]; c {
		case ' ':
			pos++
		case '\t':
			pos++
		case '[':
			tk(utOpenSq, "[")
			pos++
		case ']':
			tk(utCloseSq, "]")
			pos++
		case '(':
			tk(utOpenPar, "(")
			pos++
		case ')':
			tk(utClosePar, ")")
			pos++
		case '|':
			tk(utChoice, "|")
			pos++
		case '.':
			start := pos
			pos++
			if pos >= eof || usage[pos] != '.' {
				return nil, err("Unexpected end of usage, was expecting '..'")
			}
			pos++
			if pos >= eof || usage[pos] != '.' {
				return nil, err("Unexpected end of usage, was expecting '.'")
			}
			tkp(utRep, "...", start)
			pos++
		case '-':
			start := pos
			pos++
			if pos >= eof {
				return nil, err("Unexpected end of usage, was expecting an option name")
			}

			switch o := usage[pos]; {
			case isLetter(o):
				pos++
				for ; pos < eof; pos++ {
					ok := isLetter(usage[pos])
					if !ok {
						break
					}
				}
				typ := utShortOpt
				if pos-start > 2 {
					typ = utOptSeq
					start++
				}
				opt := usage[start:pos]
				tkp(typ, opt, start)
				if pos < eof && usage[pos] == '-' {
					return nil, err("Invalid syntax")
				}
			case o == '-':
				pos++
				if pos == eof || usage[pos] == ' ' {
					tkp(utDoubleDash, "--", start)
					continue
				}
				for pos0 := pos; pos < eof; pos++ {
					ok := isOkLongOpt(usage[pos], pos == pos0)
					if !ok {
						break
					}
				}
				opt := usage[start:pos]
				if len(opt) == 2 {
					return nil, err("Was expecting a long option name")
				}
				tkp(utLongOpt, opt, start)
			}

		case '=':
			start := pos
			pos++
			if pos >= eof || usage[pos] != '<' {
				return nil, err("Unexpected end of usage, was expecting '=<'")
			}
			closed := false
			for ; pos < eof; pos++ {
				closed = usage[pos] == '>'
				if closed {
					break
				}
			}
			if !closed {
				return nil, err("Unclosed option value")
			}
			if pos-start == 2 {
				return nil, err("Was expecting an option value")
			}
			pos++
			value := usage[start:pos]

			tkp(utOptValue, value, start)

		default:
			switch {
			case isUppercase(c):
				start := pos
				for pos = pos + 1; pos < eof; pos++ {
					if !isOkPos(usage[pos]) {
						break
					}
				}
				s := usage[start:pos]
				typ := utPos
				if s == "OPTIONS" {
					typ = utOptions
				}
				tkp(typ, s, start)
			default:
				return nil, err("Unexpected input")
			}

		}
	}

	return res, nil
}

func isLowercase(c uint8) bool {
	return c >= 'a' && c <= 'z'
}

func isUppercase(c uint8) bool {
	return c >= 'A' && c <= 'Z'
}

func isOkPos(c uint8) bool {
	return isUppercase(c) || isDigit(c) || c == '_'
}

func isLetter(c uint8) bool {
	return isLowercase(c) || isUppercase(c)
}

func isDigit(c uint8) bool {
	return c >= '0' && c <= '9'
}
func isOkLongOpt(c uint8, first bool) bool {
	return isLetter(c) || isDigit(c) || c == '_' || (!first && c == '-')
}
