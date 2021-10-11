package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func genPKCS1PrivKey() (*rsa.PrivateKey, string) {
	rsaPriv, _ := rsa.GenerateKey(rand.Reader, 4096)
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaPriv),
	}
	return rsaPriv, string(pem.EncodeToMemory(privBlock))
}

// func derivePKIXPrivKey(priv *rsa.PrivateKey) string {
// 	privBlock := &pem.Block{
// 		Type:  "RSA PRIVATE KEY",
// 		Bytes: x509.MarshalPKCS1PrivateKey(priv),
// 	}
// 	return string(pem.EncodeToMemory(privBlock))
// }

func derivePKIXPubKey(priv *rsa.PrivateKey) string {
	b, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}
	testPubKey := string(pem.EncodeToMemory(pubBlock))
	return testPubKey
}

func derivePKCS1PubKey(priv *rsa.PrivateKey) string {
	b := x509.MarshalPKCS1PublicKey(&priv.PublicKey)
	pubBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: b,
	}
	testPubKey := string(pem.EncodeToMemory(pubBlock))
	return testPubKey
}

func TestRSACrypt(t *testing.T) {
	t.Parallel()

	priv, testPrivKey := genPKCS1PrivKey()

	testdata := []struct {
		name   string
		encKey string
		decKey string
		in     []byte
	}{
		{"pkix key", derivePKIXPubKey(priv), testPrivKey, []byte("hello world")},
		{"pkcs1 key", derivePKCS1PubKey(priv), testPrivKey, []byte("hello world")},
	}

	for _, d := range testdata {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()

			enc, err := RSAEncrypt(d.encKey, d.in)
			assert.NoError(t, err)

			dec, err := RSADecrypt(d.decKey, enc)
			assert.NoError(t, err)
			assert.Equal(t, d.in, dec)
		})
	}

	t.Run("bad key", func(t *testing.T) {
		t.Parallel()

		in := []byte("hello world")
		key := "bad key"
		_, err := RSAEncrypt(key, in)
		assert.Error(t, err)

		_, err = RSADecrypt(key, in)
		assert.Error(t, err)
	})

	t.Run("empty key", func(t *testing.T) {
		t.Parallel()

		in := []byte("hello world")
		key := ""
		_, err := RSAEncrypt(key, in)
		assert.Error(t, err)

		_, err = RSADecrypt(key, in)
		assert.Error(t, err)
	})
}

func TestRSAGenerateKey(t *testing.T) {
	t.Parallel()

	_, err := RSAGenerateKey(0)
	assert.Error(t, err)

	_, err = RSAGenerateKey(12)
	assert.Error(t, err)

	key, err := RSAGenerateKey(2048)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(key),
		"-----BEGIN RSA PRIVATE KEY-----"))
	assert.True(t, strings.HasSuffix(string(key),
		"-----END RSA PRIVATE KEY-----\n"))
}

func TestRSADerivePublicKey(t *testing.T) {
	t.Parallel()

	_, err := RSADerivePublicKey(nil)
	assert.Error(t, err)

	_, err = RSADerivePublicKey([]byte(`-----BEGIN FOO-----
-----END FOO-----`))
	assert.Error(t, err)

	priv, privKey := genPKCS1PrivKey()
	expected := derivePKIXPubKey(priv)

	actual, err := RSADerivePublicKey([]byte(privKey))
	assert.NoError(t, err)
	assert.Equal(t, expected, string(actual))
}
