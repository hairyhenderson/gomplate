package funcs

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase64Encode(t *testing.T) {
	bf := &Base64Funcs{}
	assert.Equal(t, "Zm9vYmFy", must(bf.Encode("foobar")))
}

func TestBase64Decode(t *testing.T) {
	bf := &Base64Funcs{}
	assert.Equal(t, "foobar", must(bf.Decode("Zm9vYmFy")))
}

func TestBase64DecodeBytes(t *testing.T) {
	bf := &Base64Funcs{}
	out, err := bf.DecodeBytes("Zm9vYmFy")
	assert.NoError(t, err)
	assert.Equal(t, "foobar", string(out))
}

func TestToBytes(t *testing.T) {
	assert.Equal(t, []byte{0, 1, 2, 3}, toBytes([]byte{0, 1, 2, 3}))

	buf := &bytes.Buffer{}
	buf.WriteString("hi")
	assert.Equal(t, []byte("hi"), toBytes(buf))
	assert.Equal(t, []byte{}, toBytes(nil))
	assert.Equal(t, []byte("42"), toBytes(42))
}
