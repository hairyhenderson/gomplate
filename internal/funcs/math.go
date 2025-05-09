package funcs

import (
	"context"
	"fmt"
	gmath "math"
	"slices"
	"strconv"

	"github.com/hairyhenderson/gomplate/v4/conv"

	"github.com/hairyhenderson/gomplate/v4/math"
)

// CreateMathFuncs -
func CreateMathFuncs(ctx context.Context) map[string]any {
	f := map[string]any{}

	ns := &MathFuncs{ctx}
	f["math"] = func() any { return ns }

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
func (f MathFuncs) IsInt(n any) bool {
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
func (f MathFuncs) IsFloat(n any) bool {
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

// IsNum -
func (f MathFuncs) IsNum(n any) bool {
	return f.IsInt(n) || f.IsFloat(n)
}

// Abs -
func (f MathFuncs) Abs(n any) (any, error) {
	fn, err := conv.ToFloat64(n)
	if err != nil {
		return nil, fmt.Errorf("expected a number: %w", err)
	}

	m := gmath.Abs(fn)
	if f.IsInt(n) {
		return conv.ToInt64(m)
	}

	return m, nil
}

// Add -
func (f MathFuncs) Add(n ...any) (any, error) {
	if slices.ContainsFunc(n, f.IsFloat) {
		nums, err := conv.ToFloat64s(n...)
		if err != nil {
			return nil, fmt.Errorf("expected number inputs: %w", err)
		}

		var x float64
		for _, v := range nums {
			x += v
		}

		return x, nil
	}

	nums, err := conv.ToInt64s(n...)
	if err != nil {
		return nil, fmt.Errorf("expected number inputs: %w", err)
	}

	var x int64
	for _, v := range nums {
		x += v
	}

	return x, nil
}

// Mul -
func (f MathFuncs) Mul(n ...any) (any, error) {
	if slices.ContainsFunc(n, f.IsFloat) {
		nums, err := conv.ToFloat64s(n...)
		if err != nil {
			return nil, fmt.Errorf("expected number inputs: %w", err)
		}

		x := 1.
		for _, v := range nums {
			x *= v
		}

		return x, nil
	}

	nums, err := conv.ToInt64s(n...)
	if err != nil {
		return nil, fmt.Errorf("expected number inputs: %w", err)
	}

	x := int64(1)
	for _, v := range nums {
		x *= v
	}

	return x, nil
}

// Sub -
func (f MathFuncs) Sub(a, b any) (any, error) {
	if slices.ContainsFunc([]any{a, b}, f.IsFloat) {
		fa, err := conv.ToFloat64(a)
		if err != nil {
			return nil, fmt.Errorf("expected a number: %w", err)
		}

		fb, err := conv.ToFloat64(b)
		if err != nil {
			return nil, fmt.Errorf("expected a number: %w", err)
		}

		return fa - fb, nil
	}

	ia, err := conv.ToInt64(a)
	if err != nil {
		return nil, fmt.Errorf("expected a number: %w", err)
	}

	ib, err := conv.ToInt64(b)
	if err != nil {
		return nil, fmt.Errorf("expected a number: %w", err)
	}

	return ia - ib, nil
}

// Div -
func (f MathFuncs) Div(a, b any) (any, error) {
	divisor, err := conv.ToFloat64(a)
	if err != nil {
		return nil, fmt.Errorf("expected a number: %w", err)
	}

	dividend, err := conv.ToFloat64(b)
	if err != nil {
		return nil, fmt.Errorf("expected a number: %w", err)
	}

	if dividend == 0 {
		return 0, fmt.Errorf("error: division by 0")
	}

	return divisor / dividend, nil
}

// Rem -
func (f MathFuncs) Rem(a, b any) (any, error) {
	ia, err := conv.ToInt64(a)
	if err != nil {
		return nil, fmt.Errorf("expected a number: %w", err)
	}

	ib, err := conv.ToInt64(b)
	if err != nil {
		return nil, fmt.Errorf("expected a number: %w", err)
	}

	return ia % ib, nil
}

// Pow -
func (f MathFuncs) Pow(a, b any) (any, error) {
	fa, err := conv.ToFloat64(a)
	if err != nil {
		return nil, fmt.Errorf("expected a number: %w", err)
	}

	fb, err := conv.ToFloat64(b)
	if err != nil {
		return nil, fmt.Errorf("expected a number: %w", err)
	}

	r := gmath.Pow(fa, fb)
	if f.IsFloat(a) {
		return r, nil
	}

	return conv.ToInt64(r)
}

// Seq - return a sequence from `start` to `end`, in steps of `step`
// start and step are optional, and default to 1.
func (f MathFuncs) Seq(n ...any) ([]int64, error) {
	start := int64(1)
	end := int64(0)
	step := int64(1)

	var err error

	switch len(n) {
	case 1:
		end, err = conv.ToInt64(n[0])
		if err != nil {
			return nil, fmt.Errorf("expected a number: %w", err)
		}
	case 2:
		start, err = conv.ToInt64(n[0])
		if err != nil {
			return nil, fmt.Errorf("expected a number: %w", err)
		}

		end, err = conv.ToInt64(n[1])
		if err != nil {
			return nil, fmt.Errorf("expected a number: %w", err)
		}
	case 3:
		start, err = conv.ToInt64(n[0])
		if err != nil {
			return nil, fmt.Errorf("expected a number: %w", err)
		}

		end, err = conv.ToInt64(n[1])
		if err != nil {
			return nil, fmt.Errorf("expected a number: %w", err)
		}

		step, err = conv.ToInt64(n[2])
		if err != nil {
			return nil, fmt.Errorf("expected a number: %w", err)
		}
	default:
		return nil, fmt.Errorf("expected 1, 2, or 3 arguments, got %d", len(n))
	}

	return math.Seq(start, end, step), nil
}

// Max -
func (f MathFuncs) Max(a any, b ...any) (any, error) {
	if f.IsFloat(a) || slices.ContainsFunc(b, f.IsFloat) {
		m, err := conv.ToFloat64(a)
		if err != nil {
			return nil, fmt.Errorf("expected a number: %w", err)
		}

		floats, err := conv.ToFloat64s(b...)
		if err != nil {
			return nil, fmt.Errorf("expected a number: %w", err)
		}

		for _, n := range floats {
			m = gmath.Max(m, n)
		}

		return m, nil
	}

	m, err := conv.ToInt64(a)
	if err != nil {
		return nil, fmt.Errorf("expected a number: %w", err)
	}

	nums, err := conv.ToInt64s(b...)
	if err != nil {
		return nil, fmt.Errorf("expected a number: %w", err)
	}

	for _, n := range nums {
		if n > m {
			m = n
		}
	}

	return m, nil
}

// Min -
func (f MathFuncs) Min(a any, b ...any) (any, error) {
	if f.IsFloat(a) || slices.ContainsFunc(b, f.IsFloat) {
		m, err := conv.ToFloat64(a)
		if err != nil {
			return nil, fmt.Errorf("expected a number: %w", err)
		}

		floats, err := conv.ToFloat64s(b...)
		if err != nil {
			return nil, fmt.Errorf("expected a number: %w", err)
		}

		for _, n := range floats {
			m = gmath.Min(m, n)
		}
		return m, nil
	}

	m, err := conv.ToInt64(a)
	if err != nil {
		return nil, fmt.Errorf("expected a number: %w", err)
	}

	nums, err := conv.ToInt64s(b...)
	if err != nil {
		return nil, fmt.Errorf("expected a number: %w", err)
	}

	for _, n := range nums {
		if n < m {
			m = n
		}
	}
	return m, nil
}

// Ceil -
func (f MathFuncs) Ceil(n any) (any, error) {
	in, err := conv.ToFloat64(n)
	if err != nil {
		return nil, fmt.Errorf("n must be a number: %w", err)
	}

	return gmath.Ceil(in), nil
}

// Floor -
func (f MathFuncs) Floor(n any) (any, error) {
	in, err := conv.ToFloat64(n)
	if err != nil {
		return nil, fmt.Errorf("n must be a number: %w", err)
	}

	return gmath.Floor(in), nil
}

// Round -
func (f MathFuncs) Round(n any) (any, error) {
	in, err := conv.ToFloat64(n)
	if err != nil {
		return nil, fmt.Errorf("n must be a number: %w", err)
	}

	return gmath.Round(in), nil
}
