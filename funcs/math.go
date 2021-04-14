package funcs

import (
	"context"
	"fmt"
	gmath "math"
	"strconv"

	"github.com/hairyhenderson/gomplate/v3/conv"

	"github.com/hairyhenderson/gomplate/v3/math"
)

// MathNS - the math namespace
// Deprecated: don't use
func MathNS() *MathFuncs {
	return &MathFuncs{}
}

// AddMathFuncs -
// Deprecated: use CreateMathFuncs instead
func AddMathFuncs(f map[string]interface{}) {
	for k, v := range CreateMathFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateMathFuncs -
func CreateMathFuncs(ctx context.Context) map[string]interface{} {
	f := map[string]interface{}{}

	ns := &MathFuncs{ctx}
	f["math"] = func() interface{} { return ns }

	f["add"] = ns.Add
	f["sub"] = ns.Sub
	f["mul"] = ns.Mul
	f["div"] = ns.Div
	f["rem"] = ns.Rem
	f["pow"] = ns.Pow
	f["seq"] = ns.Seq
	return f
}

// MathFuncs -
type MathFuncs struct {
	ctx context.Context
}

// IsInt -
func (f MathFuncs) IsInt(n interface{}) bool {
	switch i := n.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	case string:
		_, err := strconv.ParseInt(i, 0, 64)
		return err == nil
	}
	return false
}

// IsFloat -
func (f MathFuncs) IsFloat(n interface{}) bool {
	switch i := n.(type) {
	case float32, float64:
		return true
	case string:
		_, err := strconv.ParseFloat(i, 64)
		if err != nil {
			return false
		}
		if f.IsInt(i) {
			return false
		}
		return true
	}
	return false
}

func (f MathFuncs) containsFloat(n ...interface{}) bool {
	c := false
	for _, v := range n {
		if f.IsFloat(v) {
			return true
		}
	}
	return c
}

// IsNum -
func (f MathFuncs) IsNum(n interface{}) bool {
	return f.IsInt(n) || f.IsFloat(n)
}

// Abs -
func (f MathFuncs) Abs(n interface{}) interface{} {
	m := gmath.Abs(conv.ToFloat64(n))
	if f.IsInt(n) {
		return conv.ToInt64(m)
	}
	return m
}

// Add -
func (f MathFuncs) Add(n ...interface{}) interface{} {
	if f.containsFloat(n...) {
		nums := conv.ToFloat64s(n...)
		var x float64
		for _, v := range nums {
			x += v
		}
		return x
	}
	nums := conv.ToInt64s(n...)
	var x int64
	for _, v := range nums {
		x += v
	}
	return x
}

// Mul -
func (f MathFuncs) Mul(n ...interface{}) interface{} {
	if f.containsFloat(n...) {
		nums := conv.ToFloat64s(n...)
		x := 1.
		for _, v := range nums {
			x *= v
		}
		return x
	}
	nums := conv.ToInt64s(n...)
	x := int64(1)
	for _, v := range nums {
		x *= v
	}
	return x
}

// Sub -
func (f MathFuncs) Sub(a, b interface{}) interface{} {
	if f.containsFloat(a, b) {
		return conv.ToFloat64(a) - conv.ToFloat64(b)
	}
	return conv.ToInt64(a) - conv.ToInt64(b)
}

// Div -
func (f MathFuncs) Div(a, b interface{}) (interface{}, error) {
	divisor := conv.ToFloat64(a)
	dividend := conv.ToFloat64(b)
	if dividend == 0 {
		return 0, fmt.Errorf("error: division by 0")
	}
	return divisor / dividend, nil
}

// Rem -
func (f MathFuncs) Rem(a, b interface{}) interface{} {
	return conv.ToInt64(a) % conv.ToInt64(b)
}

// Pow -
func (f MathFuncs) Pow(a, b interface{}) interface{} {
	r := gmath.Pow(conv.ToFloat64(a), conv.ToFloat64(b))
	if f.IsFloat(a) {
		return r
	}
	return conv.ToInt64(r)
}

// Seq - return a sequence from `start` to `end`, in steps of `step`
// start and step are optional, and default to 1.
func (f MathFuncs) Seq(n ...interface{}) ([]int64, error) {
	start := int64(1)
	end := int64(0)
	step := int64(1)
	if len(n) == 0 {
		return nil, fmt.Errorf("math.Seq must be given at least an 'end' value")
	}
	if len(n) == 1 {
		end = conv.ToInt64(n[0])
	}
	if len(n) == 2 {
		start = conv.ToInt64(n[0])
		end = conv.ToInt64(n[1])
	}
	if len(n) == 3 {
		start = conv.ToInt64(n[0])
		end = conv.ToInt64(n[1])
		step = conv.ToInt64(n[2])
	}
	return math.Seq(conv.ToInt64(start), conv.ToInt64(end), conv.ToInt64(step)), nil
}

// Max -
func (f MathFuncs) Max(a interface{}, b ...interface{}) (interface{}, error) {
	if f.IsFloat(a) || f.containsFloat(b...) {
		m := conv.ToFloat64(a)
		for _, n := range conv.ToFloat64s(b...) {
			m = gmath.Max(m, n)
		}
		return m, nil
	}
	m := conv.ToInt64(a)
	for _, n := range conv.ToInt64s(b...) {
		if n > m {
			m = n
		}
	}
	return m, nil
}

// Min -
func (f MathFuncs) Min(a interface{}, b ...interface{}) (interface{}, error) {
	if f.IsFloat(a) || f.containsFloat(b...) {
		m := conv.ToFloat64(a)
		for _, n := range conv.ToFloat64s(b...) {
			m = gmath.Min(m, n)
		}
		return m, nil
	}
	m := conv.ToInt64(a)
	for _, n := range conv.ToInt64s(b...) {
		if n < m {
			m = n
		}
	}
	return m, nil
}

// Ceil -
func (f MathFuncs) Ceil(n interface{}) interface{} {
	return gmath.Ceil(conv.ToFloat64(n))
}

// Floor -
func (f MathFuncs) Floor(n interface{}) interface{} {
	return gmath.Floor(conv.ToFloat64(n))
}

// Round -
func (f MathFuncs) Round(n interface{}) interface{} {
	return gmath.Round(conv.ToFloat64(n))
}
