package funcs

import (
	"context"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/hairyhenderson/gomplate/v4/conv"
	iconv "github.com/hairyhenderson/gomplate/v4/internal/conv"
	"github.com/hairyhenderson/gomplate/v4/random"
)

// CreateRandomFuncs -
func CreateRandomFuncs(ctx context.Context) map[string]interface{} {
	ns := &RandomFuncs{ctx}
	return map[string]interface{}{
		"random": func() interface{} { return ns },
	}
}

// RandomFuncs -
type RandomFuncs struct {
	ctx context.Context
}

// ASCII -
func (RandomFuncs) ASCII(count interface{}) (string, error) {
	return random.StringBounds(conv.ToInt(count), ' ', '~')
}

// Alpha -
func (RandomFuncs) Alpha(count interface{}) (string, error) {
	return random.StringRE(conv.ToInt(count), "[[:alpha:]]")
}

// AlphaNum -
func (RandomFuncs) AlphaNum(count interface{}) (string, error) {
	return random.StringRE(conv.ToInt(count), "[[:alnum:]]")
}

// String -
func (RandomFuncs) String(count interface{}, args ...interface{}) (s string, err error) {
	c := conv.ToInt(count)
	if c == 0 {
		return "", fmt.Errorf("count must be greater than 0")
	}
	m := ""
	switch len(args) {
	case 0:
		m = ""
	case 1:
		m = conv.ToString(args[0])
	case 2:
		var l, u rune
		if isString(args[0]) && isString(args[1]) {
			l, u, err = toCodePoints(args[0].(string), args[1].(string))
			if err != nil {
				return "", err
			}
		} else {
			l = rune(conv.ToInt(args[0]))
			u = rune(conv.ToInt(args[1]))
		}

		return random.StringBounds(c, l, u)
	}

	return random.StringRE(c, m)
}

func isString(s interface{}) bool {
	switch s.(type) {
	case string:
		return true
	default:
		return false
	}
}

var rlen = utf8.RuneCountInString

func toCodePoints(l, u string) (rune, rune, error) {
	// no way are these representing valid printable codepoints - we'll treat
	// them as runes
	if rlen(l) == 1 && rlen(u) == 1 {
		lower, _ := utf8.DecodeRuneInString(l)
		upper, _ := utf8.DecodeRuneInString(u)
		return lower, upper, nil
	}

	li, err := strconv.ParseInt(l, 0, 32)
	if err != nil {
		return 0, 0, err
	}
	ui, err := strconv.ParseInt(u, 0, 32)
	if err != nil {
		return 0, 0, err
	}

	return rune(li), rune(ui), nil
}

// Item -
func (RandomFuncs) Item(items interface{}) (interface{}, error) {
	i, err := iconv.InterfaceSlice(items)
	if err != nil {
		return nil, err
	}
	return random.Item(i)
}

// Number -
func (RandomFuncs) Number(args ...interface{}) (int64, error) {
	var min, max int64
	min, max = 0, 100
	switch len(args) {
	case 0:
	case 1:
		max = conv.ToInt64(args[0])
	case 2:
		min = conv.ToInt64(args[0])
		max = conv.ToInt64(args[1])
	}
	return random.Number(min, max)
}

// Float -
func (RandomFuncs) Float(args ...interface{}) (float64, error) {
	var min, max float64
	min, max = 0, 1.0
	switch len(args) {
	case 0:
	case 1:
		max = conv.ToFloat64(args[0])
	case 2:
		min = conv.ToFloat64(args[0])
		max = conv.ToFloat64(args[1])
	}
	return random.Float(min, max)
}
