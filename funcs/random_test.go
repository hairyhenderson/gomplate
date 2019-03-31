package funcs

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestASCII(t *testing.T) {
	f := &RandomFuncs{}
	s, err := f.ASCII(0)
	assert.NoError(t, err)
	assert.Empty(t, s)

	s, err = f.ASCII(100)
	assert.NoError(t, err)
	assert.Len(t, s, 100)
	assert.Regexp(t, "^[[:print:]]*$", s)
}

func TestAlpha(t *testing.T) {
	f := &RandomFuncs{}
	s, err := f.Alpha(0)
	assert.NoError(t, err)
	assert.Empty(t, s)

	s, err = f.Alpha(100)
	assert.NoError(t, err)
	assert.Len(t, s, 100)
	assert.Regexp(t, "^[[:alpha:]]*$", s)
}

func TestAlphaNum(t *testing.T) {
	f := &RandomFuncs{}
	s, err := f.AlphaNum(0)
	assert.NoError(t, err)
	assert.Empty(t, s)

	s, err = f.AlphaNum(100)
	assert.NoError(t, err)
	assert.Len(t, s, 100)
	assert.Regexp(t, "^[[:alnum:]]*$", s)
}

func TestToCodePoints(t *testing.T) {
	l, u, err := toCodePoints("a", "b")
	assert.NoError(t, err)
	assert.Equal(t, 'a', l)
	assert.Equal(t, 'b', u)

	_, _, err = toCodePoints("foo", "bar")
	assert.Error(t, err)

	_, _, err = toCodePoints("0755", "bar")
	assert.Error(t, err)

	l, u, err = toCodePoints("0xD700", "0x0001FFFF")
	assert.NoError(t, err)
	assert.Equal(t, '\ud700', l)
	assert.Equal(t, '\U0001ffff', u)

	l, u, err = toCodePoints("0011", "0777")
	assert.NoError(t, err)
	assert.Equal(t, rune(0011), l)
	assert.Equal(t, rune(0777), u)

	l, u, err = toCodePoints("♬", "♟")
	assert.NoError(t, err)
	assert.Equal(t, rune(0x266C), l)
	assert.Equal(t, '♟', u)
}

func TestString(t *testing.T) {
	f := &RandomFuncs{}
	out, err := f.String(1)
	assert.NoError(t, err)
	assert.Len(t, out, 1)

	out, err = f.String(42)
	assert.NoError(t, err)
	assert.Len(t, out, 42)

	_, err = f.String(0)
	assert.Error(t, err)

	out, err = f.String(8, "[a-z]")
	assert.NoError(t, err)
	assert.Regexp(t, "^[a-z]{8}$", out)

	out, err = f.String(10, 0x23, 0x26)
	assert.NoError(t, err)
	assert.Regexp(t, "^[#$%&]{10}$", out)

	out, err = f.String(8, '\U0001f062', '\U0001f093')
	assert.NoError(t, err)
	assert.Regexp(t, "^[🁢-🂓]{8}$", out)

	out, err = f.String(8, '\U0001f062', '\U0001f093')
	assert.NoError(t, err)
	assert.Regexp(t, "^[🁢-🂓]{8}$", out)

	out, err = f.String(8, "♚", "♟")
	assert.NoError(t, err)
	assert.Regexp(t, "^[♚-♟]{8}$", out)

	out, err = f.String(100, "♠", "♣")
	assert.NoError(t, err)
	assert.Equal(t, 100, utf8.RuneCountInString(out))
	assert.Regexp(t, "^[♠-♣]{100}$", out)
}

func TestItem(t *testing.T) {
	f := &RandomFuncs{}
	_, err := f.Item(nil)
	assert.Error(t, err)

	_, err = f.Item("foo")
	assert.Error(t, err)

	i, err := f.Item([]string{"foo"})
	assert.NoError(t, err)
	assert.Equal(t, "foo", i)

	in := []string{"foo", "bar"}
	got := ""
	for j := 0; j < 10; j++ {
		i, err = f.Item(in)
		assert.NoError(t, err)
		got += i.(string)
	}
	assert.NotEqual(t, "foofoofoofoofoofoofoofoofoofoo", got)
	assert.NotEqual(t, "barbarbarbarbarbarbarbarbarbar", got)
}

func TestNumber(t *testing.T) {
	f := &RandomFuncs{}
	n, err := f.Number()
	assert.NoError(t, err)
	assert.True(t, 0 <= n && n <= 100, n)

	_, err = f.Number(-1)
	assert.Error(t, err)

	n, err = f.Number(0)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)

	n, err = f.Number(9, 9)
	assert.NoError(t, err)
	assert.Equal(t, int64(9), n)

	n, err = f.Number(-10, -10)
	assert.NoError(t, err)
	assert.Equal(t, int64(-10), n)
}

func TestFloat(t *testing.T) {
	f := &RandomFuncs{}
	n, err := f.Float()
	assert.NoError(t, err)
	assert.InDelta(t, 0.5, n, 0.5)

	n, err = f.Float(0.5)
	assert.NoError(t, err)
	assert.InDelta(t, 0.25, n, 0.25)

	n, err = f.Float(490, 500)
	assert.NoError(t, err)
	assert.InDelta(t, 495, n, 5)

	n, err = f.Float(-500, 500)
	assert.NoError(t, err)
	assert.InDelta(t, 0, n, 500)
}
