package funcs

import (
	"context"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUUIDFuncs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateUUIDFuncs(ctx)
			actual := fmap["uuid"].(func() interface{})

			assert.Equal(t, ctx, actual().(*UUIDFuncs).ctx)
		})
	}
}

const (
	uuidV1Pattern = "^[[:xdigit:]]{8}-[[:xdigit:]]{4}-1[[:xdigit:]]{3}-[89ab][[:xdigit:]]{3}-[[:xdigit:]]{12}$"
	uuidV4Pattern = "^[[:xdigit:]]{8}-[[:xdigit:]]{4}-4[[:xdigit:]]{3}-[89ab][[:xdigit:]]{3}-[[:xdigit:]]{12}$"
)

func TestV1(t *testing.T) {
	t.Parallel()

	u := UUIDNS()
	i, err := u.V1()
	require.NoError(t, err)
	assert.Regexp(t, uuidV1Pattern, i)
}

func TestV4(t *testing.T) {
	t.Parallel()

	u := UUIDNS()
	i, err := u.V4()
	require.NoError(t, err)
	assert.Regexp(t, uuidV4Pattern, i)
}

func TestNil(t *testing.T) {
	t.Parallel()

	u := UUIDNS()
	i, err := u.Nil()
	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", i)
}

func TestIsValid(t *testing.T) {
	t.Parallel()

	u := UUIDNS()
	in := interface{}(false)
	i, err := u.IsValid(in)
	require.NoError(t, err)
	assert.False(t, i)

	in = 12345
	i, err = u.IsValid(in)
	require.NoError(t, err)
	assert.False(t, i)

	testdata := []interface{}{
		"123456781234123412341234567890ab",
		"12345678-1234-1234-1234-1234567890ab",
		"urn:uuid:12345678-1234-1234-1234-1234567890ab",
		"{12345678-1234-1234-1234-1234567890ab}",
	}

	for _, d := range testdata {
		i, err = u.IsValid(d)
		require.NoError(t, err)
		assert.True(t, i)
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	u := UUIDNS()
	in := interface{}(false)
	_, err := u.Parse(in)
	assert.Error(t, err)

	in = 12345
	_, err = u.Parse(in)
	assert.Error(t, err)

	in = "12345678-1234-1234-1234-1234567890ab"
	testdata := []interface{}{
		"123456781234123412341234567890ab",
		"12345678-1234-1234-1234-1234567890ab",
		"urn:uuid:12345678-1234-1234-1234-1234567890ab",
		must(url.Parse("urn:uuid:12345678-1234-1234-1234-1234567890ab")),
		"{12345678-1234-1234-1234-1234567890ab}",
	}

	for _, d := range testdata {
		uid, err := u.Parse(d)
		require.NoError(t, err)
		assert.Equal(t, in, uid.String())
	}
}
