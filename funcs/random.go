package funcs

import (
	"reflect"
	"strconv"
	"sync"
	"unicode/utf8"

	"github.com/hairyhenderson/gomplate/conv"
	"github.com/hairyhenderson/gomplate/random"
	"github.com/pkg/errors"
)

var (
	randomNS     *RandomFuncs
	randomNSInit sync.Once
)

// RandomNS -
func RandomNS() *RandomFuncs {
	randomNSInit.Do(func() { randomNS = &RandomFuncs{} })
	return randomNS
}

// AddRandomFuncs -
func AddRandomFuncs(f map[string]interface{}) {
	f["random"] = RandomNS
}

// RandomFuncs -
type RandomFuncs struct{}

// ASCII -
func (f *RandomFuncs) ASCII(count interface{}) (string, error) {
	return random.StringBounds(conv.ToInt(count), ' ', '~')
}

// Alpha -
func (f *RandomFuncs) Alpha(count interface{}) (string, error) {
	return random.StringRE(conv.ToInt(count), "[[:alpha:]]")
}

// AlphaNum -
func (f *RandomFuncs) AlphaNum(count interface{}) (string, error) {
	return random.StringRE(conv.ToInt(count), "[[:alnum:]]")
}

// String -
func (f *RandomFuncs) String(count interface{}, args ...interface{}) (s string, err error) {
	c := conv.ToInt(count)
	if c == 0 {
		return "", errors.New("count must be greater than 0")
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
	if rlen(l) == rlen(u) && rlen(l) == 1 {
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

func interfaceSlice(slice interface{}) ([]interface{}, error) {
	s := reflect.ValueOf(slice)
	kind := s.Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		ret := make([]interface{}, s.Len())
		for i := 0; i < s.Len(); i++ {
			ret[i] = s.Index(i).Interface()
		}
		return ret, nil
	default:
		return nil, errors.Errorf("expected an array or slice, but got a %T", s)
	}
}

// Item -
func (f *RandomFuncs) Item(items interface{}) (interface{}, error) {
	i, err := interfaceSlice(items)
	if err != nil {
		return nil, err
	}
	return random.Item(i)
}

// Number -
func (f *RandomFuncs) Number(args ...interface{}) (int64, error) {
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
func (f *RandomFuncs) Float(args ...interface{}) (float64, error) {
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
