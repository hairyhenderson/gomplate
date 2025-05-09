package crypto

import (
	"crypto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPBKDF2(t *testing.T) {
	t.Parallel()

	dk, err := PBKDF2([]byte{}, []byte{}, 0, 0, 0)
	assert.Nil(t, dk)
	require.Error(t, err)

	// IEEE 802.11i-2004 test vectors
	dk, err = PBKDF2([]byte("password"), []byte("IEEE"), 4096, 32, crypto.SHA1)
	assert.Equal(t, []byte{
		0xf4, 0x2c, 0x6f, 0xc5, 0x2d, 0xf0, 0xeb, 0xef,
		0x9e, 0xbb, 0x4b, 0x90, 0xb3, 0x8a, 0x5f, 0x90,
		0x2e, 0x83, 0xfe, 0x1b, 0x13, 0x5a, 0x70, 0xe2,
		0x3a, 0xed, 0x76, 0x2e, 0x97, 0x10, 0xa1, 0x2e,
	}, dk)
	require.NoError(t, err)

	dk, err = PBKDF2([]byte("ThisIsAPassword"), []byte("ThisIsASSID"), 4096, 32, crypto.SHA1)
	assert.Equal(t, []byte{
		0x0d, 0xc0, 0xd6, 0xeb, 0x90, 0x55, 0x5e, 0xd6,
		0x41, 0x97, 0x56, 0xb9, 0xa1, 0x5e, 0xc3, 0xe3,
		0x20, 0x9b, 0x63, 0xdf, 0x70, 0x7d, 0xd5, 0x08,
		0xd1, 0x45, 0x81, 0xf8, 0x98, 0x27, 0x21, 0xaf,
	}, dk)
	require.NoError(t, err)

	dk, err = PBKDF2([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), []byte("ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"), 4096, 32, crypto.SHA1)
	assert.Equal(t, []byte{
		0xbe, 0xcb, 0x93, 0x86, 0x6b, 0xb8, 0xc3, 0x83,
		0x2c, 0xb7, 0x77, 0xc2, 0xf5, 0x59, 0x80, 0x7c,
		0x8c, 0x59, 0xaf, 0xcb, 0x6e, 0xae, 0x73, 0x48,
		0x85, 0x00, 0x13, 0x00, 0xa9, 0x81, 0xcc, 0x62,
	}, dk)
	require.NoError(t, err)

	// some longer hash functions
	dk, err = PBKDF2([]byte("password"), []byte("IEEE"), 4096, 64, crypto.SHA512)
	assert.Equal(t, []byte{
		0xc1, 0x6f, 0x4c, 0xb6, 0xd0, 0x3e, 0x23, 0x61,
		0x43, 0x99, 0xde, 0xe5, 0xe7, 0xf6, 0x76, 0xfb,
		0x1d, 0xa0, 0xeb, 0x94, 0x71, 0xb6, 0xa7, 0x4a,
		0x6c, 0x5b, 0xc9, 0x34, 0xc6, 0xec, 0x7d, 0x2a,
		0xb7, 0x02, 0x8f, 0xbb, 0x10, 0x00, 0xb1, 0xbe,
		0xb9, 0x7f, 0x17, 0x64, 0x60, 0x45, 0xd8, 0x14,
		0x47, 0x92, 0x35, 0x2f, 0x66, 0x76, 0xd1, 0x3b,
		0x20, 0xa4, 0xc0, 0x37, 0x54, 0x90, 0x3d, 0x7e,
	}, dk)
	require.NoError(t, err)
}

func TestStrToHash(t *testing.T) {
	t.Parallel()

	h, err := StrToHash("foo")
	assert.Zero(t, h)
	require.Error(t, err)
	h, err = StrToHash("SHA-1")
	assert.Equal(t, crypto.SHA1, h)
	require.NoError(t, err)
	h, err = StrToHash("SHA224")
	assert.Equal(t, crypto.SHA224, h)
	require.NoError(t, err)
	h, err = StrToHash("SHA-256")
	assert.Equal(t, crypto.SHA256, h)
	require.NoError(t, err)
	h, err = StrToHash("SHA384")
	assert.Equal(t, crypto.SHA384, h)
	require.NoError(t, err)
	h, err = StrToHash("SHA-512")
	assert.Equal(t, crypto.SHA512, h)
	require.NoError(t, err)
	h, err = StrToHash("SHA-512/224")
	assert.Equal(t, crypto.SHA512_224, h)
	require.NoError(t, err)
	h, err = StrToHash("SHA512/256")
	assert.Equal(t, crypto.SHA512_256, h)
	require.NoError(t, err)
}
