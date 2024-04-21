package crypto

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEd25519GenerateKey(t *testing.T) {
	key, err := Ed25519GenerateKey()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(key),
		"-----BEGIN PRIVATE KEY-----"))
	assert.True(t, strings.HasSuffix(string(key),
		"-----END PRIVATE KEY-----\n"))
}

func TestEd25519DerivePublicKey(t *testing.T) {
	_, err := Ed25519DerivePublicKey(nil)
	require.Error(t, err)
	_, err = Ed25519DerivePublicKey([]byte(`-----BEGIN FOO-----
	-----END FOO-----`))
	require.Error(t, err)

	priv, err := Ed25519GenerateKey()
	require.NoError(t, err)
	pub, err := Ed25519DerivePublicKey(priv)
	require.NoError(t, err)
	block, _ := pem.Decode(pub)
	assert.NotNil(t, block)
	secret, err := ed25519DecodeFromPEM(priv)
	require.NoError(t, err)
	p, err := x509.ParsePKIXPublicKey(block.Bytes)
	require.NoError(t, err)
	pubKey, ok := p.(ed25519.PublicKey)
	assert.True(t, ok)
	assert.Equal(t, fmt.Sprintf("%x", secret.Public()), fmt.Sprintf("%x", p))
	msg := []byte("ed25519")
	sig := ed25519.Sign(secret, msg) // Panics
	assert.True(t, ed25519.Verify(pubKey, msg, sig))
}
