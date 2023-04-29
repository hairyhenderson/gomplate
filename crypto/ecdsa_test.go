package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func genECDSAPrivKey() (*ecdsa.PrivateKey, string) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	der, _ := x509.MarshalECPrivateKey(priv)
	privBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	}
	return priv, string(pem.EncodeToMemory(privBlock))
}

func deriveECPubkey(priv *ecdsa.PrivateKey) string {
	b, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}
	testPubKey := string(pem.EncodeToMemory(pubBlock))
	return testPubKey
}

func TestECDSAGenerateKey(t *testing.T) {
	key, err := ECDSAGenerateKey(elliptic.P224())
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(key),
		"-----BEGIN EC PRIVATE KEY-----"))
	assert.True(t, strings.HasSuffix(string(key),
		"-----END EC PRIVATE KEY-----\n"))

	key, err = ECDSAGenerateKey(elliptic.P256())
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(key),
		"-----BEGIN EC PRIVATE KEY-----"))
	assert.True(t, strings.HasSuffix(string(key),
		"-----END EC PRIVATE KEY-----\n"))

	key, err = ECDSAGenerateKey(elliptic.P384())
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(key),
		"-----BEGIN EC PRIVATE KEY-----"))
	assert.True(t, strings.HasSuffix(string(key),
		"-----END EC PRIVATE KEY-----\n"))

	key, err = ECDSAGenerateKey(elliptic.P521())
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(key),
		"-----BEGIN EC PRIVATE KEY-----"))
	assert.True(t, strings.HasSuffix(string(key),
		"-----END EC PRIVATE KEY-----\n"))
}

func TestECDSADerivePublicKey(t *testing.T) {
	_, err := ECDSADerivePublicKey(nil)
	assert.Error(t, err)

	_, err = ECDSADerivePublicKey([]byte(`-----BEGIN FOO-----
	-----END FOO-----`))
	assert.Error(t, err)

	priv, privKey := genECDSAPrivKey()
	expected := deriveECPubkey(priv)

	actual, err := ECDSADerivePublicKey([]byte(privKey))
	require.NoError(t, err)
	assert.Equal(t, expected, string(actual))
}
