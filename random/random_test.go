package random

import (
	"math"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestMatchChars(t *testing.T) {
	in := "[a-g]"
	expected := []rune("abcdefg")
	out, err := matchChars(in)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)

	in = "[a-zA-Z0-9_.-]"
	expected = []rune(defaultSet)
	out, err = matchChars(in)
	assert.NoError(t, err)
	assert.Equal(t, expected, out)

	in = "[[:alpha:]]"
	expected = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	out, err = matchChars(in)
	assert.NoError(t, err)
	assert.Equal(t, expected, out)
}

func TestStringRE(t *testing.T) {
	r, err := StringRE(15, "[\\p{Yi}[:alnum:]]")
	assert.NoError(t, err)
	assert.Equal(t, 15, utf8.RuneCountInString(r))

	_, err = StringRE(1, "[bogus")
	assert.Error(t, err)
}

func TestStringBounds(t *testing.T) {
	_, err := StringBounds(15, 0, 19)
	assert.Error(t, err)

	// surrogate range isn't valid, should error
	_, err = StringBounds(15, 0xd800, 0xdfff)
	assert.Error(t, err)

	r, err := StringBounds(1, 'a', 'a')
	assert.NoError(t, err)
	assert.Equal(t, "a", r)

	r, err = StringBounds(99, 'a', 'b')
	assert.NoError(t, err)
	assert.Regexp(t, "^[a-b]+$", r)

	r, err = StringBounds(100, 0x0020, 0x007f)
	assert.NoError(t, err)
	assert.Regexp(t, "^[\u0020-\u007f]*$", r)

	// only 🂱 (\U0001F0B1) in this range is "graphic"
	r, err = StringBounds(8, 0x0001f0af, 0x0001f0b1)
	assert.NoError(t, err)
	assert.Regexp(t, "^🂱🂱🂱🂱🂱🂱🂱🂱$", r)
}

func TestItem(t *testing.T) {
	_, err := Item(nil)
	assert.Error(t, err)

	i, err := Item([]interface{}{"foo"})
	assert.NoError(t, err)
	assert.Equal(t, "foo", i)

	in := []interface{}{"foo", "bar"}
	got := ""
	for j := 0; j < 10; j++ {
		i, err = Item(in)
		assert.NoError(t, err)
		got += i.(string)
	}
	assert.NotEqual(t, "foofoofoofoofoofoofoofoofoofoo", got)
	assert.NotEqual(t, "barbarbarbarbarbarbarbarbarbar", got)
}

func TestNumber(t *testing.T) {
	_, err := Number(0, -1)
	assert.Error(t, err)
	_, err = Number(0, math.MaxInt64)
	assert.Error(t, err)
	_, err = Number(math.MinInt64, 0)
	assert.Error(t, err)

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
		assert.NoError(t, err)
		assert.InDelta(t, d.expected, n, d.delta)
	}
}

func TestFloat(t *testing.T) {
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
		assert.NoError(t, err)
		assert.InDelta(t, d.expected, n, d.delta)
	}
}
