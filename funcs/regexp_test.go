package funcs

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateReFuncs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateReFuncs(ctx)
			actual := fmap["regexp"].(func() interface{})

			assert.Same(t, ctx, actual().(*ReFuncs).ctx)
		})
	}
}

func TestReplace(t *testing.T) {
	t.Parallel()

	re := &ReFuncs{}
	assert.Equal(t, "hello world", re.Replace("i", "ello", "hi world"))
}

func TestMatch(t *testing.T) {
	t.Parallel()

	re := &ReFuncs{}
	assert.True(t, re.Match(`i\ `, "hi world"))
}

func TestFind(t *testing.T) {
	t.Parallel()

	re := &ReFuncs{}
	f, err := re.Find(`[a-z]+`, `foo bar baz`)
	assert.NoError(t, err)
	assert.Equal(t, "foo", f)

	_, err = re.Find(`[a-`, "")
	assert.Error(t, err)

	f, err = re.Find("4", 42)
	assert.NoError(t, err)
	assert.Equal(t, "4", f)

	f, err = re.Find(false, 42)
	assert.NoError(t, err)
	assert.Equal(t, "", f)
}

func TestFindAll(t *testing.T) {
	t.Parallel()

	re := &ReFuncs{}
	f, err := re.FindAll(`[a-z]+`, `foo bar baz`)
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"foo", "bar", "baz"}, f)

	f, err = re.FindAll(`[a-z]+`, -1, `foo bar baz`)
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"foo", "bar", "baz"}, f)

	_, err = re.FindAll(`[a-`, "")
	assert.Error(t, err)

	_, err = re.FindAll("")
	assert.Error(t, err)

	_, err = re.FindAll("", "", "", "")
	assert.Error(t, err)

	f, err = re.FindAll(`[a-z]+`, 0, `foo bar baz`)
	assert.NoError(t, err)
	assert.Nil(t, f)

	f, err = re.FindAll(`[a-z]+`, 2, `foo bar baz`)
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"foo", "bar"}, f)

	f, err = re.FindAll(`[a-z]+`, 14, `foo bar baz`)
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"foo", "bar", "baz"}, f)

	f, err = re.FindAll(`qux`, `foo bar baz`)
	assert.NoError(t, err)
	assert.Nil(t, f)
}

func TestSplit(t *testing.T) {
	t.Parallel()

	re := &ReFuncs{}
	f, err := re.Split(` `, `foo bar baz`)
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"foo", "bar", "baz"}, f)

	f, err = re.Split(`\s+`, -1, `foo  bar baz`)
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"foo", "bar", "baz"}, f)

	_, err = re.Split(`[a-`, "")
	assert.Error(t, err)

	_, err = re.Split("")
	assert.Error(t, err)

	_, err = re.Split("", "", "", "")
	assert.Error(t, err)

	f, err = re.Split(` `, 0, `foo bar baz`)
	assert.NoError(t, err)
	assert.Nil(t, f)

	f, err = re.Split(`\s+`, 2, `foo bar baz`)
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"foo", "bar baz"}, f)

	f, err = re.Split(`\s`, 14, `foo  bar baz`)
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"foo", "", "bar", "baz"}, f)

	f, err = re.Split(`[\s,.]`, 14, `foo bar.baz,qux`)
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"foo", "bar", "baz", "qux"}, f)
}

func TestReplaceLiteral(t *testing.T) {
	t.Parallel()

	re := &ReFuncs{}
	r, err := re.ReplaceLiteral("i", "ello$1", "hi world")
	assert.NoError(t, err)
	assert.Equal(t, "hello$1 world", r)
}
