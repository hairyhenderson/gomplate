package funcs

import (
	"context"
	"strconv"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRandomFuncs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateRandomFuncs(ctx)
			actual := fmap["random"].(func() interface{})

			assert.Equal(t, ctx, actual().(*RandomFuncs).ctx)
		})
	}
}

func TestASCII(t *testing.T) {
	t.Parallel()

	f := RandomFuncs{}
	s, err := f.ASCII(0)
	require.NoError(t, err)
	assert.Empty(t, s)

	s, err = f.ASCII(100)
	require.NoError(t, err)
	assert.Len(t, s, 100)
	assert.Regexp(t, "^[[:print:]]*$", s)
}

func TestAlpha(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping slow test")
	}

	f := RandomFuncs{}
	s, err := f.Alpha(0)
	require.NoError(t, err)
	assert.Empty(t, s)

	s, err = f.Alpha(100)
	require.NoError(t, err)
	assert.Len(t, s, 100)
	assert.Regexp(t, "^[[:alpha:]]*$", s)
}

func TestAlphaNum(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping slow test")
	}

	f := RandomFuncs{}
	s, err := f.AlphaNum(0)
	require.NoError(t, err)
	assert.Empty(t, s)

	s, err = f.AlphaNum(100)
	require.NoError(t, err)
	assert.Len(t, s, 100)
	assert.Regexp(t, "^[[:alnum:]]*$", s)
}

func TestToCodePoints(t *testing.T) {
	t.Parallel()

	l, u, err := toCodePoints("a", "b")
	require.NoError(t, err)
	assert.Equal(t, 'a', l)
	assert.Equal(t, 'b', u)

	_, _, err = toCodePoints("foo", "bar")
	assert.Error(t, err)

	_, _, err = toCodePoints("0755", "bar")
	assert.Error(t, err)

	l, u, err = toCodePoints("0xD700", "0x0001FFFF")
	require.NoError(t, err)
	assert.Equal(t, '\ud700', l)
	assert.Equal(t, '\U0001ffff', u)

	l, u, err = toCodePoints("0011", "0777")
	require.NoError(t, err)
	assert.Equal(t, rune(0011), l)
	assert.Equal(t, rune(0777), u)

	l, u, err = toCodePoints("♬", "♟")
	require.NoError(t, err)
	assert.Equal(t, rune(0x266C), l)
	assert.Equal(t, '♟', u)
}

func TestString(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping slow test")
	}

	f := RandomFuncs{}
	out, err := f.String(1)
	require.NoError(t, err)
	assert.Len(t, out, 1)

	out, err = f.String(42)
	require.NoError(t, err)
	assert.Len(t, out, 42)

	_, err = f.String(0)
	assert.Error(t, err)

	out, err = f.String(8, "[a-z]")
	require.NoError(t, err)
	assert.Regexp(t, "^[a-z]{8}$", out)

	out, err = f.String(10, 0x23, 0x26)
	require.NoError(t, err)
	assert.Regexp(t, "^[#$%&]{10}$", out)

	out, err = f.String(8, '\U0001f062', '\U0001f093')
	require.NoError(t, err)
	assert.Regexp(t, "^[🁢-🂓]{8}$", out)

	out, err = f.String(8, '\U0001f062', '\U0001f093')
	require.NoError(t, err)
	assert.Regexp(t, "^[🁢-🂓]{8}$", out)

	out, err = f.String(8, "♚", "♟")
	require.NoError(t, err)
	assert.Regexp(t, "^[♚-♟]{8}$", out)

	out, err = f.String(100, "♠", "♣")
	require.NoError(t, err)
	assert.Equal(t, 100, utf8.RuneCountInString(out))
	assert.Regexp(t, "^[♠-♣]{100}$", out)
}

func TestItem(t *testing.T) {
	t.Parallel()

	f := RandomFuncs{}
	_, err := f.Item(nil)
	assert.Error(t, err)

	_, err = f.Item("foo")
	assert.Error(t, err)

	i, err := f.Item([]string{"foo"})
	require.NoError(t, err)
	assert.Equal(t, "foo", i)

	in := []string{"foo", "bar"}
	got := ""
	for j := 0; j < 10; j++ {
		i, err = f.Item(in)
		require.NoError(t, err)
		got += i.(string)
	}
	assert.NotEqual(t, "foofoofoofoofoofoofoofoofoofoo", got)
	assert.NotEqual(t, "barbarbarbarbarbarbarbarbarbar", got)
}

func TestNumber(t *testing.T) {
	t.Parallel()

	f := RandomFuncs{}
	n, err := f.Number()
	require.NoError(t, err)
	assert.True(t, 0 <= n && n <= 100, n)

	_, err = f.Number(-1)
	assert.Error(t, err)

	n, err = f.Number(0)
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)

	n, err = f.Number(9, 9)
	require.NoError(t, err)
	assert.Equal(t, int64(9), n)

	n, err = f.Number(-10, -10)
	require.NoError(t, err)
	assert.Equal(t, int64(-10), n)
}

func TestFloat(t *testing.T) {
	t.Parallel()

	f := RandomFuncs{}
	n, err := f.Float()
	require.NoError(t, err)
	assert.InDelta(t, 0.5, n, 0.5)

	n, err = f.Float(0.5)
	require.NoError(t, err)
	assert.InDelta(t, 0.25, n, 0.25)

	n, err = f.Float(490, 500)
	require.NoError(t, err)
	assert.InDelta(t, 495, n, 5)

	n, err = f.Float(-500, 500)
	require.NoError(t, err)
	assert.InDelta(t, 0, n, 500)
}
