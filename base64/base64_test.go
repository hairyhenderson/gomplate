package base64

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func must(r interface{}, err error) interface{} {
	if err != nil {
		return err
	}
	return r
}

func TestEncode(t *testing.T) {
	assert.Equal(t, "", must(Encode([]byte(""))))
	assert.Equal(t, "Zg==", must(Encode([]byte("f"))))
	assert.Equal(t, "Zm8=", must(Encode([]byte("fo"))))
	assert.Equal(t, "Zm9v", must(Encode([]byte("foo"))))
	assert.Equal(t, "Zm9vYg==", must(Encode([]byte("foob"))))
	assert.Equal(t, "Zm9vYmE=", must(Encode([]byte("fooba"))))
	assert.Equal(t, "Zm9vYmFy", must(Encode([]byte("foobar"))))
	assert.Equal(t, "A+B/", must(Encode([]byte{0x03, 0xe0, 0x7f})))
}

func TestDecode(t *testing.T) {
	assert.Equal(t, []byte(""), must(Decode("")))
	assert.Equal(t, []byte("f"), must(Decode("Zg==")))
	assert.Equal(t, []byte("fo"), must(Decode("Zm8=")))
	assert.Equal(t, []byte("foo"), must(Decode("Zm9v")))
	assert.Equal(t, []byte("foob"), must(Decode("Zm9vYg==")))
	assert.Equal(t, []byte("fooba"), must(Decode("Zm9vYmE=")))
	assert.Equal(t, []byte("foobar"), must(Decode("Zm9vYmFy")))
	assert.Equal(t, []byte{0x03, 0xe0, 0x7f}, must(Decode("A+B/")))
	assert.Equal(t, []byte{0x03, 0xe0, 0x7f}, must(Decode("A-B_")))

	_, err := Decode("b.o.g.u.s")
	assert.Error(t, err)
}
