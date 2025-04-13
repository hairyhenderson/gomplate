package funcs

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateReFuncs(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateReFuncs(ctx)
			actual := fmap["regexp"].(func() any)

			assert.Equal(t, ctx, actual().(*ReFuncs).ctx)
		})
	}
}

func TestReplace(t *testing.T) {
	t.Parallel()

	re := &ReFuncs{}

	actual, err := re.Replace("i", "ello", "hi world")
	require.NoError(t, err)
	assert.Equal(t, "hello world", actual)
}

func TestMatch(t *testing.T) {
	t.Parallel()

	re := &ReFuncs{}

	actual, err := re.Match(`i\ `, "hi world")
	require.NoError(t, err)
	assert.True(t, actual)
}

func TestFind(t *testing.T) {
	t.Parallel()

	re := &ReFuncs{}
	f, err := re.Find(`[a-z]+`, `foo bar baz`)
	require.NoError(t, err)
	assert.Equal(t, "foo", f)

	_, err = re.Find(`[a-`, "")
	require.Error(t, err)

	f, err = re.Find("4", 42)
	require.NoError(t, err)
	assert.Equal(t, "4", f)

	f, err = re.Find(false, 42)
	require.NoError(t, err)
	assert.Empty(t, f)
}

func TestFindAll(t *testing.T) {
	t.Parallel()

	re := &ReFuncs{}
	f, err := re.FindAll(`[a-z]+`, `foo bar baz`)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar", "baz"}, f)

	f, err = re.FindAll(`[a-z]+`, -1, `foo bar baz`)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar", "baz"}, f)

	_, err = re.FindAll(`[a-`, "")
	require.Error(t, err)

	_, err = re.FindAll("")
	require.Error(t, err)

	_, err = re.FindAll("", "", "", "")
	require.Error(t, err)

	f, err = re.FindAll(`[a-z]+`, 0, `foo bar baz`)
	require.NoError(t, err)
	assert.Nil(t, f)

	f, err = re.FindAll(`[a-z]+`, 2, `foo bar baz`)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar"}, f)

	f, err = re.FindAll(`[a-z]+`, 14, `foo bar baz`)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar", "baz"}, f)

	f, err = re.FindAll(`qux`, `foo bar baz`)
	require.NoError(t, err)
	assert.Nil(t, f)
}

func TestSplit(t *testing.T) {
	t.Parallel()

	re := &ReFuncs{}
	f, err := re.Split(` `, `foo bar baz`)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar", "baz"}, f)

	f, err = re.Split(`\s+`, -1, `foo  bar baz`)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar", "baz"}, f)

	_, err = re.Split(`[a-`, "")
	require.Error(t, err)

	_, err = re.Split("")
	require.Error(t, err)

	_, err = re.Split("", "", "", "")
	require.Error(t, err)

	f, err = re.Split(` `, 0, `foo bar baz`)
	require.NoError(t, err)
	assert.Nil(t, f)

	f, err = re.Split(`\s+`, 2, `foo bar baz`)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar baz"}, f)

	f, err = re.Split(`\s`, 14, `foo  bar baz`)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo", "", "bar", "baz"}, f)

	f, err = re.Split(`[\s,.]`, 14, `foo bar.baz,qux`)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar", "baz", "qux"}, f)
}

func TestReplaceLiteral(t *testing.T) {
	t.Parallel()

	re := &ReFuncs{}
	r, err := re.ReplaceLiteral("i", "ello$1", "hi world")
	require.NoError(t, err)
	assert.Equal(t, "hello$1 world", r)
}
