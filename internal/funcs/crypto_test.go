package funcs

import (
	"context"
	"encoding/base64"
	"strconv"
	"strings"
	"testing"

	"github.com/hairyhenderson/gomplate/v5/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCryptoFuncs(t *testing.T) {
	for i := range 10 {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateCryptoFuncs(ctx)
			actual := fmap["crypto"].(func() any)

			assert.Equal(t, ctx, actual().(*CryptoFuncs).ctx)
		})
	}
}

func testCryptoNS() *CryptoFuncs {
	return &CryptoFuncs{ctx: config.SetExperimental(context.Background())}
}

func TestPBKDF2(t *testing.T) {
	t.Parallel()

	c := testCryptoNS()
	dk, err := c.PBKDF2("password", []byte("IEEE"), "4096", 32)
	assert.Equal(t, "f42c6fc52df0ebef9ebb4b90b38a5f902e83fe1b135a70e23aed762e9710a12e", dk)
	require.NoError(t, err)

	dk, err = c.PBKDF2([]byte("password"), "IEEE", 4096, "64", "SHA-512")
	assert.Equal(t, "c16f4cb6d03e23614399dee5e7f676fb1da0eb9471b6a74a6c5bc934c6ec7d2ab7028fbb1000b1beb97f17646045d8144792352f6676d13b20a4c03754903d7e", dk)
	require.NoError(t, err)

	_, err = c.PBKDF2(nil, nil, nil, nil, "bogus")
	require.Error(t, err)
}

func TestWPAPSK(t *testing.T) {
	t.Parallel()

	c := testCryptoNS()
	dk, err := c.WPAPSK("password", "MySSID")
	assert.Equal(t, "3a98def84b11644a17ebcc9b17955d2360ce8b8a85b8a78413fc551d722a84e7", dk)
	require.NoError(t, err)
}

func TestSHA(t *testing.T) {
	t.Parallel()

	in := "abc"
	sha1 := "a9993e364706816aba3e25717850c26c9cd0d89d"
	sha224 := "23097d223405d8228642a477bda255b32aadbce4bda0b3f7e36c9da7"
	sha256 := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	sha384 := "cb00753f45a35e8bb5a03d699ac65007272c32ab0eded1631a8b605a43ff5bed8086072ba1e7cc2358baeca134c825a7"
	sha512 := "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f"
	sha512_224 := "4634270f707b6a54daae7530460842e20e37ed265ceee9a43e8924aa"
	sha512_256 := "53048e2681941ef99b2e29b76b4c7dabe4c2d0c634fc6d46e0e2f13107e7af23"
	c := testCryptoNS()
	assert.Equal(t, sha1, c.SHA1(in))
	assert.Equal(t, sha224, c.SHA224(in))
	assert.Equal(t, sha256, c.SHA256(in))
	assert.Equal(t, sha384, c.SHA384(in))
	assert.Equal(t, sha512, c.SHA512(in))
	assert.Equal(t, sha512_224, c.SHA512_224(in))
	assert.Equal(t, sha512_256, c.SHA512_256(in))
}

func TestBcrypt(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping slow test")
	}

	in := "foo"
	c := testCryptoNS()

	t.Run("no arg default", func(t *testing.T) {
		t.Parallel()

		actual, err := c.Bcrypt(in)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(actual, "$2a$10$"))
	})

	t.Run("cost less than min", func(t *testing.T) {
		t.Parallel()

		actual, err := c.Bcrypt(0, in)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(actual, "$2a$10$"))
	})

	t.Run("cost equal to min", func(t *testing.T) {
		t.Parallel()

		actual, err := c.Bcrypt(4, in)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(actual, "$2a$04$"))
	})

	t.Run("no args errors", func(t *testing.T) {
		t.Parallel()

		_, err := c.Bcrypt()
		require.Error(t, err)
	})
}

func TestRSAGenerateKey(t *testing.T) {
	t.Parallel()

	c := testCryptoNS()
	_, err := c.RSAGenerateKey(0)
	require.Error(t, err)

	_, err = c.RSAGenerateKey(0, "foo", true)
	require.Error(t, err)

	key, err := c.RSAGenerateKey(2048)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(key,
		"-----BEGIN RSA PRIVATE KEY-----"))
	assert.True(t, strings.HasSuffix(key,
		"-----END RSA PRIVATE KEY-----\n"))
}

func TestECDSAGenerateKey(t *testing.T) {
	c := testCryptoNS()
	_, err := c.ECDSAGenerateKey("")
	require.Error(t, err)

	_, err = c.ECDSAGenerateKey(0, "P-999", true)
	require.Error(t, err)

	key, err := c.ECDSAGenerateKey("P-256")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(key,
		"-----BEGIN EC PRIVATE KEY-----"))
	assert.True(t, strings.HasSuffix(key,
		"-----END EC PRIVATE KEY-----\n"))
}

func TestECDSADerivePublicKey(t *testing.T) {
	c := testCryptoNS()

	_, err := c.ECDSADerivePublicKey("")
	require.Error(t, err)

	key, _ := c.ECDSAGenerateKey("P-256")
	pub, err := c.ECDSADerivePublicKey(key)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(pub,
		"-----BEGIN PUBLIC KEY-----"))
	assert.True(t, strings.HasSuffix(pub,
		"-----END PUBLIC KEY-----\n"))
}

func TestEd25519GenerateKey(t *testing.T) {
	c := testCryptoNS()
	key, err := c.Ed25519GenerateKey()
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(key,
		"-----BEGIN PRIVATE KEY-----"))
	assert.True(t, strings.HasSuffix(key,
		"-----END PRIVATE KEY-----\n"))
}

func TestEd25519GenerateKeyFromSeed(t *testing.T) {
	c := testCryptoNS()
	enc := ""
	seed := ""
	_, err := c.Ed25519GenerateKeyFromSeed(enc, seed)
	require.Error(t, err)

	enc = "base64"
	seed = "0000000000000000000000000000000" // 31 bytes, instead of wanted 32.
	_, err = c.Ed25519GenerateKeyFromSeed(enc, seed)
	require.Error(t, err)

	seed += "0" // 32 bytes.
	b64seed := base64.StdEncoding.EncodeToString([]byte(seed))
	key, err := c.Ed25519GenerateKeyFromSeed(enc, b64seed)
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(key,
		"-----BEGIN PRIVATE KEY-----"))
	assert.True(t, strings.HasSuffix(key,
		"-----END PRIVATE KEY-----\n"))
}

func TestEd25519DerivePublicKey(t *testing.T) {
	c := testCryptoNS()

	_, err := c.Ed25519DerivePublicKey("")
	require.Error(t, err)

	key, _ := c.Ed25519GenerateKey()
	pub, err := c.Ed25519DerivePublicKey(key)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(pub,
		"-----BEGIN PUBLIC KEY-----"))
	assert.True(t, strings.HasSuffix(pub,
		"-----END PUBLIC KEY-----\n"))
}

func TestRSACrypt(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping slow test")
	}

	c := testCryptoNS()
	key, err := c.RSAGenerateKey()
	require.NoError(t, err)
	pub, err := c.RSADerivePublicKey(key)
	require.NoError(t, err)

	in := "hello world"
	enc, err := c.RSAEncrypt(pub, in)
	require.NoError(t, err)

	dec, err := c.RSADecrypt(key, enc)
	require.NoError(t, err)
	assert.Equal(t, in, dec)

	b, err := c.RSADecryptBytes(key, enc)
	require.NoError(t, err)
	assert.Equal(t, dec, string(b))
}

func TestAESCrypt(t *testing.T) {
	c := testCryptoNS()
	key := "0123456789012345"
	in := "hello world"

	_, err := c.EncryptAES(key, 1, 2, 3, 4)
	require.Error(t, err)

	_, err = c.DecryptAES(key, 1, 2, 3, 4)
	require.Error(t, err)

	enc, err := c.EncryptAES(key, in)
	require.NoError(t, err)

	dec, err := c.DecryptAES(key, enc)
	require.NoError(t, err)
	assert.Equal(t, in, dec)

	b, err := c.DecryptAESBytes(key, enc)
	require.NoError(t, err)
	assert.Equal(t, dec, string(b))

	enc, err = c.EncryptAES(key, 128, in)
	require.NoError(t, err)

	dec, err = c.DecryptAES(key, 128, enc)
	require.NoError(t, err)
	assert.Equal(t, in, dec)

	b, err = c.DecryptAESBytes(key, 128, enc)
	require.NoError(t, err)
	assert.Equal(t, dec, string(b))
}
