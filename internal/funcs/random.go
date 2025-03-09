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
func CreateRandomFuncs(ctx context.Context) map[string]any {
	ns := &RandomFuncs{ctx}
	return map[string]any{
		"random": func() any { return ns },
	}
}

// RandomFuncs -
type RandomFuncs struct {
	ctx context.Context
}

// ASCII -
func (RandomFuncs) ASCII(count any) (string, error) {
	n, err := conv.ToInt(count)
	if err != nil {
		return "", fmt.Errorf("count must be an integer: %w", err)
	}

	return random.StringBounds(n, ' ', '~')
}

// Alpha -
func (RandomFuncs) Alpha(count any) (string, error) {
	n, err := conv.ToInt(count)
	if err != nil {
		return "", fmt.Errorf("count must be an integer: %w", err)
	}

	return random.StringRE(n, "[[:alpha:]]")
}

// AlphaNum -
func (RandomFuncs) AlphaNum(count any) (string, error) {
	n, err := conv.ToInt(count)
	if err != nil {
		return "", fmt.Errorf("count must be an integer: %w", err)
	}

	return random.StringRE(n, "[[:alnum:]]")
}

// String -
func (RandomFuncs) String(count any, args ...any) (string, error) {
	c, err := conv.ToInt(count)
	if err != nil {
		return "", fmt.Errorf("count must be an integer: %w", err)
	}

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
			nl, err := conv.ToInt(args[0])
			if err != nil {
				return "", fmt.Errorf("lower must be an integer: %w", err)
			}

			nu, err := conv.ToInt(args[1])
			if err != nil {
				return "", fmt.Errorf("upper must be an integer: %w", err)
			}

			l, u = rune(nl), rune(nu)
		}

		return random.StringBounds(c, l, u)
	}

	return random.StringRE(c, m)
}

func isString(s any) bool {
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
func (RandomFuncs) Item(items any) (any, error) {
	i, err := iconv.InterfaceSlice(items)
	if err != nil {
		return nil, err
	}
	return random.Item(i)
}

// Number -
func (RandomFuncs) Number(args ...any) (int64, error) {
	var nMin, nMax int64
	nMin, nMax = 0, 100

	var err error

	switch len(args) {
	case 0:
	case 1:
		nMax, err = conv.ToInt64(args[0])
		if err != nil {
			return 0, fmt.Errorf("max must be a number: %w", err)
		}
	case 2:
		nMin, err = conv.ToInt64(args[0])
		if err != nil {
			return 0, fmt.Errorf("min must be a number: %w", err)
		}

		nMax, err = conv.ToInt64(args[1])
		if err != nil {
			return 0, fmt.Errorf("max must be a number: %w", err)
		}
	}

	return random.Number(nMin, nMax)
}

// Float -
func (RandomFuncs) Float(args ...any) (float64, error) {
	var nMin, nMax float64
	nMin, nMax = 0, 1.0

	var err error

	switch len(args) {
	case 0:
	case 1:
		nMax, err = conv.ToFloat64(args[0])
		if err != nil {
			return 0, fmt.Errorf("max must be a number: %w", err)
		}
	case 2:
		nMin, err = conv.ToFloat64(args[0])
		if err != nil {
			return 0, fmt.Errorf("min must be a number: %w", err)
		}

		nMax, err = conv.ToFloat64(args[1])
		if err != nil {
			return 0, fmt.Errorf("max must be a number: %w", err)
		}
	}

	return random.Float(nMin, nMax)
}
