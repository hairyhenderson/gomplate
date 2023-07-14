package funcs

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCryptoFuncs(t *testing.T) {
	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateCryptoFuncs(ctx)
			actual := fmap["crypto"].(func() interface{})

			assert.Same(t, ctx, actual().(*CryptoFuncs).ctx)
		})
	}
}

func testCryptoNS() *CryptoFuncs {
	return &CryptoFuncs{}
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
