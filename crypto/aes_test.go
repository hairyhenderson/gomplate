package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
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

// encryptRawCBC encrypts already-padded content with AES-CBC, prefixing the
// IV, so that tests can build ciphertexts with arbitrary (invalid) padding.
func encryptRawCBC(t *testing.T, key, iv, padded []byte) []byte {
	t.Helper()

	block, err := aes.NewCipher(key)
	require.NoError(t, err)

	out := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, padded)

	return append(append([]byte{}, iv...), out...)
}

func TestDecryptAESCBCInvalidInput(t *testing.T) {
	key := bytes.Repeat([]byte{'k'}, 32)
	iv := bytes.Repeat([]byte{'i'}, aes.BlockSize)

	valid, err := EncryptAESCBC(key, []byte("some content"))
	require.NoError(t, err)

	block := func(b byte) []byte { return bytes.Repeat([]byte{b}, aes.BlockSize) }

	// the last block is padded with the right number of bytes, but one of
	// them has the wrong value
	badPadContent := append(block('x')[:aes.BlockSize-4:aes.BlockSize-4], 4, 3, 4, 4)

	testdata := map[string][]byte{
		"shorter than the IV":         valid[:aes.BlockSize-1],
		"just the IV":                 valid[:aes.BlockSize],
		"partial ciphertext block":    valid[:len(valid)-1],
		"partial block after the IV":  valid[:aes.BlockSize+3],
		"zero padding length":         encryptRawCBC(t, key, iv, append(block('x')[:aes.BlockSize-1:aes.BlockSize-1], 0)),
		"padding longer than a block": encryptRawCBC(t, key, iv, append(block('x')[:aes.BlockSize-1:aes.BlockSize-1], aes.BlockSize+1)),
		"mismatched padding bytes":    encryptRawCBC(t, key, iv, badPadContent),
	}

	for name, in := range testdata {
		t.Run(name, func(t *testing.T) {
			require.NotPanics(t, func() {
				out, err := DecryptAESCBC(key, in)
				require.Error(t, err)
				assert.Nil(t, out)
			})
		})
	}

	// the full-block padding written by EncryptAESCBC is still fine
	out, err := DecryptAESCBC(key, encryptRawCBC(t, key, iv, append(block('x'), block(aes.BlockSize)...)))
	require.NoError(t, err)
	assert.Equal(t, block('x'), out)
}
