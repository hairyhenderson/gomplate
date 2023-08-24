package funcs

import (
	"context"
	"fmt"
	gmath "math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMathFuncs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateMathFuncs(ctx)
			actual := fmap["math"].(func() interface{})

			assert.Equal(t, ctx, actual().(*MathFuncs).ctx)
		})
	}
}

func TestAdd(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	assert.Equal(t, int64(12), m.Add(1, 1, 2, 3, 5))
	assert.Equal(t, int64(2), m.Add(1, 1))
	assert.Equal(t, int64(1), m.Add(1))
	assert.Equal(t, int64(0), m.Add(-5, 5))
	assert.InDelta(t, float64(5.1), m.Add(4.9, "0.2"), 0.000000001)
}

func TestMul(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	assert.Equal(t, int64(30), m.Mul(1, 1, 2, 3, 5))
	assert.Equal(t, int64(1), m.Mul(1, 1))
	assert.Equal(t, int64(1), m.Mul(1))
	assert.Equal(t, int64(-25), m.Mul("-5", 5))
	assert.Equal(t, int64(28), m.Mul(14, "2"))
	assert.Equal(t, float64(0.5), m.Mul("-1", -0.5))
}

func TestSub(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	assert.Equal(t, int64(0), m.Sub(1, 1))
	assert.Equal(t, int64(-10), m.Sub(-5, 5))
	assert.Equal(t, int64(-41), m.Sub(true, "42"))
	assert.InDelta(t, -5.3, m.Sub(10, 15.3), 0.000000000000001)
}

func mustDiv(a, b interface{}) interface{} {
	m := MathFuncs{}
	r, err := m.Div(a, b)
	if err != nil {
		return -1
	}
	return r
}

func TestDiv(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	_, err := m.Div(1, 0)
	assert.Error(t, err)
	assert.Equal(t, 1., mustDiv(1, 1))
	assert.Equal(t, -1., mustDiv(-5, 5))
	assert.Equal(t, 1./42, mustDiv(true, "42"))
	assert.InDelta(t, 0.5, mustDiv(1, 2), 1e-12)
}

func TestRem(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	assert.Equal(t, int64(0), m.Rem(1, 1))
	assert.Equal(t, int64(2), m.Rem(5, 3.0))
}

func TestPow(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	assert.Equal(t, int64(4), m.Pow(2, "2"))
	assert.Equal(t, 2.25, m.Pow(1.5, 2))
}

func mustSeq(t *testing.T, n ...interface{}) []int64 {
	m := MathFuncs{}
	s, err := m.Seq(n...)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func TestSeq(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	assert.EqualValues(t, []int64{0, 1, 2, 3}, mustSeq(t, 0, 3))
	assert.EqualValues(t, []int64{1, 0}, mustSeq(t, 0))
	assert.EqualValues(t, []int64{0, 2, 4}, mustSeq(t, 0, 4, 2))
	assert.EqualValues(t, []int64{0, 2, 4}, mustSeq(t, 0, 5, 2))
	assert.EqualValues(t, []int64{0}, mustSeq(t, 0, 5, 8))
	_, err := m.Seq()
	assert.Error(t, err)
}

func TestIsIntFloatNum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in      interface{}
		isInt   bool
		isFloat bool
	}{
		{0, true, false},
		{1, true, false},
		{-1, true, false},
		{uint(42), true, false},
		{uint8(255), true, false},
		{uint16(42), true, false},
		{uint32(42), true, false},
		{uint64(42), true, false},
		{int(42), true, false},
		{int8(127), true, false},
		{int16(42), true, false},
		{int32(42), true, false},
		{int64(42), true, false},
		{float32(18.3), false, true},
		{float64(18.3), false, true},
		{1.5, false, true},
		{-18.6, false, true},
		{"42", true, false},
		{"052", true, false},
		{"0xff", true, false},
		{"-42", true, false},
		{"-0", true, false},
		{"3.14", false, true},
		{"-3.14", false, true},
		{"0.00", false, true},
		{"NaN", false, true},
		{"-Inf", false, true},
		{"+Inf", false, true},
		{"", false, false},
		{"foo", false, false},
		{nil, false, false},
		{true, false, false},
	}
	m := MathFuncs{}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%T(%#v)", tt.in, tt.in), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.isInt, m.IsInt(tt.in))
			assert.Equal(t, tt.isFloat, m.IsFloat(tt.in))
			assert.Equal(t, tt.isInt || tt.isFloat, m.IsNum(tt.in))
		})
	}
}

func BenchmarkIsFloat(b *testing.B) {
	data := []interface{}{
		0, 1, -1, uint(42), uint8(255), uint16(42), uint32(42), uint64(42), int(42), int8(127), int16(42), int32(42), int64(42), float32(18.3), float64(18.3), 1.5, -18.6, "42", "052", "0xff", "-42", "-0", "3.14", "-3.14", "0.00", "NaN", "-Inf", "+Inf", "", "foo", nil, true,
	}
	m := MathFuncs{}
	for _, n := range data {
		n := n
		b.Run(fmt.Sprintf("%T(%v)", n, n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				m.IsFloat(n)
			}
		})
	}
}

func TestMax(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	data := []struct {
		expected interface{}
		n        []interface{}
	}{
		{int64(0), []interface{}{nil}},
		{int64(0), []interface{}{0}},
		{int64(0), []interface{}{"not a number"}},
		{int64(1), []interface{}{1}},
		{int64(-1), []interface{}{-1}},
		{int64(1), []interface{}{-1, 0, 1}},
		{3.9, []interface{}{3.14, 3, 3.9}},
		{int64(255), []interface{}{"14", "0xff", -5}},
	}
	for _, d := range data {
		d := d
		t.Run(fmt.Sprintf("%v==%v", d.n, d.expected), func(t *testing.T) {
			t.Parallel()

			var actual interface{}
			if len(d.n) == 1 {
				actual, _ = m.Max(d.n[0])
			} else {
				actual, _ = m.Max(d.n[0], d.n[1:]...)
			}
			assert.Equal(t, d.expected, actual)
		})
	}
}

func TestMin(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	data := []struct {
		expected interface{}
		n        []interface{}
	}{
		{int64(0), []interface{}{nil}},
		{int64(0), []interface{}{0}},
		{int64(0), []interface{}{"not a number"}},
		{int64(1), []interface{}{1}},
		{int64(-1), []interface{}{-1}},
		{int64(-1), []interface{}{-1, 0, 1}},
		{3., []interface{}{3.14, 3, 3.9}},
		{int64(-5), []interface{}{"14", "0xff", -5}},
	}
	for _, d := range data {
		d := d
		t.Run(fmt.Sprintf("%v==%v", d.n, d.expected), func(t *testing.T) {
			t.Parallel()

			var actual interface{}
			if len(d.n) == 1 {
				actual, _ = m.Min(d.n[0])
			} else {
				actual, _ = m.Min(d.n[0], d.n[1:]...)
			}
			assert.Equal(t, d.expected, actual)
		})
	}
}

func TestContainsFloat(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	data := []struct {
		n        []interface{}
		expected bool
	}{
		{[]interface{}{nil}, false},
		{[]interface{}{0}, false},
		{[]interface{}{"not a number"}, false},
		{[]interface{}{1}, false},
		{[]interface{}{-1}, false},
		{[]interface{}{-1, 0, 1}, false},
		{[]interface{}{3.14, 3, 3.9}, true},
		{[]interface{}{"14", "0xff", -5}, false},
		{[]interface{}{"14.8", "0xff", -5}, true},
		{[]interface{}{"-Inf", 2}, true},
		{[]interface{}{"NaN"}, true},
	}
	for _, d := range data {
		d := d
		t.Run(fmt.Sprintf("%v==%v", d.n, d.expected), func(t *testing.T) {
			t.Parallel()

			if d.expected {
				assert.True(t, m.containsFloat(d.n...))
			} else {
				assert.False(t, m.containsFloat(d.n...))
			}
		})
	}
}

func TestCeil(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	data := []struct {
		n interface{}
		a float64
	}{
		{"", 0.},
		{nil, 0.},
		{"Inf", gmath.Inf(1)},
		{0, 0.},
		{4.99, 5.},
		{42.1, 43},
		{-1.9, -1},
	}
	for _, d := range data {
		d := d
		t.Run(fmt.Sprintf("%v==%v", d.n, d.a), func(t *testing.T) {
			t.Parallel()

			assert.InDelta(t, d.a, m.Ceil(d.n), 1e-12)
		})
	}
}

func TestFloor(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	data := []struct {
		n interface{}
		a float64
	}{
		{"", 0.},
		{nil, 0.},
		{"Inf", gmath.Inf(1)},
		{0, 0.},
		{4.99, 4.},
		{42.1, 42},
		{-1.9, -2.},
	}
	for _, d := range data {
		d := d
		t.Run(fmt.Sprintf("%v==%v", d.n, d.a), func(t *testing.T) {
			t.Parallel()

			assert.InDelta(t, d.a, m.Floor(d.n), 1e-12)
		})
	}
}

func TestRound(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	data := []struct {
		n interface{}
		a float64
	}{
		{"", 0.},
		{nil, 0.},
		{"Inf", gmath.Inf(1)},
		{0, 0.},
		{4.99, 5},
		{42.1, 42},
		{-1.9, -2.},
		{3.5, 4},
		{-3.5, -4},
		{4.5, 5},
		{-4.5, -5},
	}
	for _, d := range data {
		d := d
		t.Run(fmt.Sprintf("%v==%v", d.n, d.a), func(t *testing.T) {
			t.Parallel()

			assert.InDelta(t, d.a, m.Round(d.n), 1e-12)
		})
	}
}

func TestAbs(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	data := []struct {
		n interface{}
		a interface{}
	}{
		{"", 0.},
		{nil, 0.},
		{"-Inf", gmath.Inf(1)},
		{0, int64(0)},
		{0., 0.},
		{gmath.Copysign(0, -1), 0.},
		{3.14, 3.14},
		{-1.9, 1.9},
		{2, int64(2)},
		{-2, int64(2)},
	}
	for _, d := range data {
		d := d
		t.Run(fmt.Sprintf("%#v==%v", d.n, d.a), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, d.a, m.Abs(d.n))
		})
	}
}
