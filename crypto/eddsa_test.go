package crypto

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEDDSAGenerateKey(t *testing.T) {
	key, err := EDDSAGenerateKey()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(key),
		"-----BEGIN PRIVATE KEY-----"))
	assert.True(t, strings.HasSuffix(string(key),
		"-----END PRIVATE KEY-----\n"))
}

func TestEDDSADerivePublicKey(t *testing.T) {
	_, err := EDDSADerivePublicKey(nil)
	assert.Error(t, err)
	_, err = EDDSADerivePublicKey([]byte(`-----BEGIN FOO-----
	-----END FOO-----`))
	assert.Error(t, err)

	priv, err := EDDSAGenerateKey()
	require.NoError(t, err)
	pub, err := EDDSADerivePublicKey(priv)
	require.NoError(t, err)
	block, _ := pem.Decode(pub)
	assert.True(t, block != nil)
	secret, err := edDSADecodeFromPEM(priv)
	require.NoError(t, err)
	p, err := x509.ParsePKIXPublicKey(block.Bytes)
	require.NoError(t, err)
	assert.True(t, fmt.Sprintf("%x", p) == fmt.Sprintf("%x", secret.Public()))
}
