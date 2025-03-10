package random

import (
	"math"
	"strconv"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchChars(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test")
	}

	t.Parallel()

	testdata := []struct {
		in       string
		expected string
	}{
		{"[a-g]", "abcdefg"},
		{"[a-zA-Z0-9_.-]", defaultSet},
		{"[[:alpha:]]", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"},
	}

	for i, d := range testdata {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			out, err := matchChars(d.in)
			require.NoError(t, err)
			assert.EqualValues(t, d.expected, out)
		})
	}
}

func TestStringRE(t *testing.T) {
	t.Parallel()

	r, err := StringRE(15, "[\\p{Yi}[:alnum:]]")
	require.NoError(t, err)
	assert.Equal(t, 15, utf8.RuneCountInString(r))

	_, err = StringRE(1, "[bogus")
	require.Error(t, err)
}

func TestStringBounds(t *testing.T) {
	t.Parallel()

	_, err := StringBounds(15, 0, 19)
	require.Error(t, err)

	// surrogate range isn't valid, should error
	_, err = StringBounds(15, 0xd800, 0xdfff)
	require.Error(t, err)

	r, err := StringBounds(1, 'a', 'a')
	require.NoError(t, err)
	assert.Equal(t, "a", r)

	r, err = StringBounds(99, 'a', 'b')
	require.NoError(t, err)
	assert.Regexp(t, "^[a-b]+$", r)

	r, err = StringBounds(100, 0x0020, 0x007f)
	require.NoError(t, err)
	assert.Regexp(t, "^[\u0020-\u007f]*$", r)

	// only 🂱 (\U0001F0B1) in this range is "graphic"
	r, err = StringBounds(8, 0x0001f0af, 0x0001f0b1)
	require.NoError(t, err)
	assert.Regexp(t, "^🂱🂱🂱🂱🂱🂱🂱🂱$", r)
}

func TestItem(t *testing.T) {
	t.Parallel()

	_, err := Item(nil)
	require.Error(t, err)

	i, err := Item([]any{"foo"})
	require.NoError(t, err)
	assert.Equal(t, "foo", i)

	in := []any{"foo", "bar"}
	got := ""
	for range 10 {
		i, err = Item(in)
		require.NoError(t, err)
		got += i.(string)
	}
	assert.NotEqual(t, "foofoofoofoofoofoofoofoofoofoo", got)
	assert.NotEqual(t, "barbarbarbarbarbarbarbarbarbar", got)
}

func TestNumber(t *testing.T) {
	t.Parallel()

	_, err := Number(0, -1)
	require.Error(t, err)
	_, err = Number(0, math.MaxInt64)
	require.Error(t, err)
	_, err = Number(math.MinInt64, 0)
	require.Error(t, err)

	testdata := []struct {
		min, max, expected int64
		delta              float64
	}{
		{0, 100, 50, 50},
		{0, 0, 0, 0},
		{9, 9, 9, 0},
		{-10, -10, -10, 0},
		{-10, -0, -5, 5},
	}
	for _, d := range testdata {
		n, err := Number(d.min, d.max)
		require.NoError(t, err)
		assert.InDelta(t, d.expected, n, d.delta)
	}
}

func TestFloat(t *testing.T) {
	t.Parallel()

	testdata := []struct {
		min, max, expected float64
		delta              float64
	}{
		{0, 1.0, 0.5, 0.5},
		{0, 0.5, 0.25, 0.25},
		{490, 500, 495, 5},
		{-500, 500, 0, 500},
		{0, math.MaxFloat64, math.MaxFloat64 / 2, math.MaxFloat64 / 2},
	}

	for _, d := range testdata {
		n, err := Float(d.min, d.max)
		require.NoError(t, err)
		assert.InDelta(t, d.expected, n, d.delta)
	}
}
