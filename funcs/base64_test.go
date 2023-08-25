package funcs

import (
	"bytes"
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBase64Funcs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateBase64Funcs(ctx)
			actual := fmap["base64"].(func() interface{})

			assert.Equal(t, ctx, actual().(*Base64Funcs).ctx)
		})
	}
}

func TestBase64Encode(t *testing.T) {
	t.Parallel()

	bf := &Base64Funcs{}
	assert.Equal(t, "Zm9vYmFy", must(bf.Encode("foobar")))
}

func TestBase64Decode(t *testing.T) {
	t.Parallel()

	bf := &Base64Funcs{}
	assert.Equal(t, "foobar", must(bf.Decode("Zm9vYmFy")))
}

func TestBase64DecodeBytes(t *testing.T) {
	t.Parallel()

	bf := &Base64Funcs{}
	out, err := bf.DecodeBytes("Zm9vYmFy")
	require.NoError(t, err)
	assert.Equal(t, "foobar", string(out))
}

func TestToBytes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []byte{0, 1, 2, 3}, toBytes([]byte{0, 1, 2, 3}))

	buf := &bytes.Buffer{}
	buf.WriteString("hi")
	assert.Equal(t, []byte("hi"), toBytes(buf))
	assert.Equal(t, []byte{}, toBytes(nil))
	assert.Equal(t, []byte("42"), toBytes(42))
}
