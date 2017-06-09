package base64

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	assert.Equal(t, "", Encode([]byte("")))
	assert.Equal(t, "Zg==", Encode([]byte("f")))
	assert.Equal(t, "Zm8=", Encode([]byte("fo")))
	assert.Equal(t, "Zm9v", Encode([]byte("foo")))
	assert.Equal(t, "Zm9vYg==", Encode([]byte("foob")))
	assert.Equal(t, "Zm9vYmE=", Encode([]byte("fooba")))
	assert.Equal(t, "Zm9vYmFy", Encode([]byte("foobar")))
}

func TestDecode(t *testing.T) {
	assert.Equal(t, []byte(""), Decode(""))
	assert.Equal(t, []byte("f"), Decode("Zg=="))
	assert.Equal(t, []byte("fo"), Decode("Zm8="))
	assert.Equal(t, []byte("foo"), Decode("Zm9v"))
	assert.Equal(t, []byte("foob"), Decode("Zm9vYg=="))
	assert.Equal(t, []byte("fooba"), Decode("Zm9vYmE="))
	assert.Equal(t, []byte("foobar"), Decode("Zm9vYmFy"))
}
