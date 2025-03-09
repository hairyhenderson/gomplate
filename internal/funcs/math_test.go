package funcs

import (
	"context"
	"fmt"
	gmath "math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateMathFuncs(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateMathFuncs(ctx)
			actual := fmap["math"].(func() any)

			assert.Equal(t, ctx, actual().(*MathFuncs).ctx)
		})
	}
}

func TestAdd(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}

	actual, err := m.Add(1, 1, 2, 3, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(12), actual)

	actual, err = m.Add(1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), actual)

	actual, err = m.Add(1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), actual)

	actual, err = m.Add(-5, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(0), actual)

	actual, err = m.Add(4.9, "0.2")
	require.NoError(t, err)
	assert.InEpsilon(t, float64(5.1), actual, 1e-12)
}

func TestMul(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}

	actual, err := m.Mul(1, 1, 2, 3, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(30), actual)

	actual, err = m.Mul(1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), actual)

	actual, err = m.Mul(1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), actual)

	actual, err = m.Mul("-5", 5)
	require.NoError(t, err)
	assert.Equal(t, int64(-25), actual)

	actual, err = m.Mul(14, "2")
	require.NoError(t, err)
	assert.Equal(t, int64(28), actual)

	actual, err = m.Mul("-1", -0.5)
	require.NoError(t, err)
	assert.InEpsilon(t, float64(0.5), actual, 1e-12)
}

func TestSub(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}

	actual, err := m.Sub(1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(0), actual)

	actual, err = m.Sub(-5, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(-10), actual)

	actual, err = m.Sub(true, "42")
	require.NoError(t, err)
	assert.Equal(t, int64(-41), actual)

	actual, err = m.Sub(10, 15.3)
	require.NoError(t, err)
	assert.InEpsilon(t, -5.3, actual, 1e-12)
}

func mustDiv(a, b any) any {
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
	require.Error(t, err)
	assert.InEpsilon(t, 1., mustDiv(1, 1), 1e-12)
	assert.InEpsilon(t, -1., mustDiv(-5, 5), 1e-12)
	assert.InEpsilon(t, 1./42, mustDiv(true, "42"), 1e-12)
	assert.InEpsilon(t, 0.5, mustDiv(1, 2), 1e-12)
}

func TestRem(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}

	actual, err := m.Rem(1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(0), actual)

	actual, err = m.Rem(5, 3.0)
	require.NoError(t, err)
	assert.Equal(t, int64(2), actual)
}

func TestPow(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}

	actual, err := m.Pow(2, "2")
	require.NoError(t, err)
	assert.Equal(t, int64(4), actual)

	actual, err = m.Pow(1.5, 2)
	require.NoError(t, err)
	assert.InEpsilon(t, 2.25, actual, 1e-12)
}

func mustSeq(t *testing.T, n ...any) []int64 {
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
	require.Error(t, err)
}

func TestIsIntFloatNum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in      any
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
		t.Run(fmt.Sprintf("%T(%#v)", tt.in, tt.in), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.isInt, m.IsInt(tt.in))
			assert.Equal(t, tt.isFloat, m.IsFloat(tt.in))
			assert.Equal(t, tt.isInt || tt.isFloat, m.IsNum(tt.in))
		})
	}
}

func BenchmarkIsFloat(b *testing.B) {
	data := []any{
		0, 1, -1, uint(42), uint8(255), uint16(42), uint32(42), uint64(42), int(42), int8(127), int16(42), int32(42), int64(42), float32(18.3), float64(18.3), 1.5, -18.6, "42", "052", "0xff", "-42", "-0", "3.14", "-3.14", "0.00", "NaN", "-Inf", "+Inf", "", "foo", nil, true,
	}

	m := MathFuncs{}
	for _, n := range data {
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
		expected any
		n        []any
	}{
		{int64(0), []any{0}},
		{int64(1), []any{1}},
		{int64(-1), []any{-1}},
		{int64(1), []any{-1, 0, 1}},
		{3.9, []any{3.14, 3, 3.9}},
		{int64(255), []any{"14", "0xff", -5}},
	}
	for _, d := range data {
		t.Run(fmt.Sprintf("%v==%v", d.n, d.expected), func(t *testing.T) {
			t.Parallel()

			var actual any
			if len(d.n) == 1 {
				actual, _ = m.Max(d.n[0])
			} else {
				actual, _ = m.Max(d.n[0], d.n[1:]...)
			}
			assert.Equal(t, d.expected, actual)
		})
	}

	t.Run("error cases", func(t *testing.T) {
		t.Parallel()

		_, err := m.Max("foo")
		require.Error(t, err)

		_, err = m.Max(nil)
		require.Error(t, err)

		_, err = m.Max("")
		require.Error(t, err)
	})
}

func TestMin(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	data := []struct {
		expected any
		n        []any
	}{
		{int64(0), []any{0}},
		{int64(1), []any{1}},
		{int64(-1), []any{-1}},
		{int64(-1), []any{-1, 0, 1}},
		{3., []any{3.14, 3, 3.9}},
		{int64(-5), []any{"14", "0xff", -5}},
	}
	for _, d := range data {
		t.Run(fmt.Sprintf("%v==%v", d.n, d.expected), func(t *testing.T) {
			t.Parallel()

			var actual any
			if len(d.n) == 1 {
				actual, _ = m.Min(d.n[0])
			} else {
				actual, _ = m.Min(d.n[0], d.n[1:]...)
			}

			assert.Equal(t, d.expected, actual)
		})
	}

	t.Run("error cases", func(t *testing.T) {
		t.Parallel()

		_, err := m.Min("foo")
		require.Error(t, err)

		_, err = m.Min(nil)
		require.Error(t, err)

		_, err = m.Min("")
		require.Error(t, err)
	})
}

func TestContainsFloat(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	data := []struct {
		n        []any
		expected bool
	}{
		{[]any{nil}, false},
		{[]any{0}, false},
		{[]any{"not a number"}, false},
		{[]any{1}, false},
		{[]any{-1}, false},
		{[]any{-1, 0, 1}, false},
		{[]any{3.14, 3, 3.9}, true},
		{[]any{"14", "0xff", -5}, false},
		{[]any{"14.8", "0xff", -5}, true},
		{[]any{"-Inf", 2}, true},
		{[]any{"NaN"}, true},
	}
	for _, d := range data {
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
		n any
		a float64
	}{
		{"Inf", gmath.Inf(1)},
		{0, 0.},
		{4.99, 5.},
		{42.1, 43},
		{-1.9, -1},
	}
	for _, d := range data {
		t.Run(fmt.Sprintf("%v==%v", d.n, d.a), func(t *testing.T) {
			t.Parallel()

			actual, err := m.Ceil(d.n)
			require.NoError(t, err)
			assert.InDelta(t, d.a, actual, 1e-12)
		})
	}

	t.Run("error cases", func(t *testing.T) {
		t.Parallel()

		_, err := m.Ceil("foo")
		require.Error(t, err)

		_, err = m.Ceil(nil)
		require.Error(t, err)

		_, err = m.Ceil("")
		require.Error(t, err)
	})
}

func TestFloor(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	data := []struct {
		n any
		a float64
	}{
		{"Inf", gmath.Inf(1)},
		{0, 0.},
		{4.99, 4.},
		{42.1, 42},
		{-1.9, -2.},
	}
	for _, d := range data {
		t.Run(fmt.Sprintf("%v==%v", d.n, d.a), func(t *testing.T) {
			t.Parallel()

			actual, err := m.Floor(d.n)
			require.NoError(t, err)
			assert.InDelta(t, d.a, actual, 1e-12)
		})
	}

	t.Run("error cases", func(t *testing.T) {
		t.Parallel()

		_, err := m.Floor("foo")
		require.Error(t, err)

		_, err = m.Floor(nil)
		require.Error(t, err)

		_, err = m.Floor("")
		require.Error(t, err)
	})
}

func TestRound(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	data := []struct {
		n any
		a float64
	}{
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
		t.Run(fmt.Sprintf("%v==%v", d.n, d.a), func(t *testing.T) {
			t.Parallel()

			actual, err := m.Round(d.n)
			require.NoError(t, err)
			assert.InDelta(t, d.a, actual, 1e-12)
		})
	}

	t.Run("error cases", func(t *testing.T) {
		t.Parallel()

		_, err := m.Round("foo")
		require.Error(t, err)

		_, err = m.Round(nil)
		require.Error(t, err)

		_, err = m.Round("")
		require.Error(t, err)
	})
}

func TestAbs(t *testing.T) {
	t.Parallel()

	m := MathFuncs{}
	data := []struct {
		n any
		a any
	}{
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
		t.Run(fmt.Sprintf("%#v==%v", d.n, d.a), func(t *testing.T) {
			t.Parallel()

			actual, err := m.Abs(d.n)
			require.NoError(t, err)
			assert.Equal(t, d.a, actual)
		})
	}

	t.Run("error cases", func(t *testing.T) {
		t.Parallel()

		_, err := m.Abs("foo")
		require.Error(t, err)

		_, err = m.Abs(nil)
		require.Error(t, err)

		_, err = m.Abs("")
		require.Error(t, err)
	})
}
