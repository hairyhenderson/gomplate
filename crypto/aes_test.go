package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptAESCBC(t *testing.T) {
	// empty key is invalid
	_, err := EncryptAESCBC([]byte{}, []byte("foo"))
	require.Error(t, err)

	// wrong-length keys are invalid
	_, err = EncryptAESCBC(bytes.Repeat([]byte{'a'}, 1), []byte("foo"))
	require.Error(t, err)

	_, err = EncryptAESCBC(bytes.Repeat([]byte{'a'}, 15), []byte("foo"))
	require.Error(t, err)

	key := make([]byte, 32)
	copy(key, []byte("password"))

	// empty content is a pass-through
	out, err := EncryptAESCBC(key, []byte{})
	require.NoError(t, err)
	assert.Equal(t, []byte{}, out)

	testdata := [][]byte{
		bytes.Repeat([]byte{'a'}, 1),
		bytes.Repeat([]byte{'a'}, 15),
		bytes.Repeat([]byte{'a'}, 16),
		bytes.Repeat([]byte{'a'}, 31),
		bytes.Repeat([]byte{'a'}, 32),
	}

	for _, d := range testdata {
		out, err = EncryptAESCBC(key, d)
		require.NoError(t, err)
		assert.NotEqual(t, d, out)

		out, err = DecryptAESCBC(key, out)
		require.NoError(t, err)
		assert.Equal(t, d, out)
	}

	// 128-bit key
	key = bytes.Repeat([]byte{'a'}, 16)
	out, err = EncryptAESCBC(key, []byte("foo"))
	require.NoError(t, err)
	assert.NotEqual(t, []byte("foo"), out)

	out, err = DecryptAESCBC(key, out)
	require.NoError(t, err)
	assert.Equal(t, []byte("foo"), out)

	// 192-bit key
	key = bytes.Repeat([]byte{'a'}, 24)
	out, err = EncryptAESCBC(key, []byte("foo"))
	require.NoError(t, err)
	assert.NotEqual(t, []byte("foo"), out)

	out, err = DecryptAESCBC(key, out)
	require.NoError(t, err)
	assert.Equal(t, []byte("foo"), out)
}
