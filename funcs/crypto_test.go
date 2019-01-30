package funcs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPBKDF2(t *testing.T) {
	c := CryptoNS()
	dk, err := cryptoNS.PBKDF2("password", []byte("IEEE"), "4096", 32)
	assert.Equal(t, "f42c6fc52df0ebef9ebb4b90b38a5f902e83fe1b135a70e23aed762e9710a12e", dk)
	assert.NoError(t, err)

	dk, err = c.PBKDF2([]byte("password"), "IEEE", 4096, "64", "SHA-512")
	assert.Equal(t, "c16f4cb6d03e23614399dee5e7f676fb1da0eb9471b6a74a6c5bc934c6ec7d2ab7028fbb1000b1beb97f17646045d8144792352f6676d13b20a4c03754903d7e", dk)
	assert.NoError(t, err)

	_, err = c.PBKDF2(nil, nil, nil, nil, "bogus")
	assert.Error(t, err)
}

func TestWPAPSK(t *testing.T) {
	dk, err := cryptoNS.WPAPSK("password", "MySSID")
	assert.Equal(t, "3a98def84b11644a17ebcc9b17955d2360ce8b8a85b8a78413fc551d722a84e7", dk)
	assert.NoError(t, err)
}

func TestSHA(t *testing.T) {
	in := "abc"
	sha1 := "a9993e364706816aba3e25717850c26c9cd0d89d"
	sha224 := "23097d223405d8228642a477bda255b32aadbce4bda0b3f7e36c9da7"
	sha256 := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	sha384 := "cb00753f45a35e8bb5a03d699ac65007272c32ab0eded1631a8b605a43ff5bed8086072ba1e7cc2358baeca134c825a7"
	sha512 := "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f"
	sha512_224 := "4634270f707b6a54daae7530460842e20e37ed265ceee9a43e8924aa"
	sha512_256 := "53048e2681941ef99b2e29b76b4c7dabe4c2d0c634fc6d46e0e2f13107e7af23"
	c := CryptoNS()
	assert.Equal(t, sha1, c.SHA1(in))
	assert.Equal(t, sha224, c.SHA224(in))
	assert.Equal(t, sha256, c.SHA256(in))
	assert.Equal(t, sha384, c.SHA384(in))
	assert.Equal(t, sha512, c.SHA512(in))
	assert.Equal(t, sha512_224, c.SHA512_224(in))
	assert.Equal(t, sha512_256, c.SHA512_256(in))
}

func TestBcrypt(t *testing.T) {
	in := "foo"
	c := CryptoNS()
	actual, err := c.Bcrypt(in)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(actual, "$2a$10$"))

	actual, err = c.Bcrypt(0, in)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(actual, "$2a$10$"))

	actual, err = c.Bcrypt(4, in)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(actual, "$2a$04$"))

	_, err = c.Bcrypt()
	assert.Error(t, err)
}
