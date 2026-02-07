package integration

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"gotest.tools/v3/fs"
)

func genTestKeys() (string, string) {
	rsaPriv, _ := rsa.GenerateKey(rand.Reader, 4096)
	rsaPub := rsaPriv.PublicKey
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaPriv),
	}
	testPrivKey := string(pem.EncodeToMemory(privBlock))

	b, _ := x509.MarshalPKIXPublicKey(&rsaPub)
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}
	testPubKey := string(pem.EncodeToMemory(pubBlock))
	return testPrivKey, testPubKey
}

func setupCryptoTest(t *testing.T) *fs.Dir {
	testPrivKey, testPubKey := genTestKeys()

	tmpDir := fs.NewDir(t, "gomplate-inttests",
		fs.WithFile("testPrivKey", testPrivKey),
		fs.WithFile("testPubKey", testPubKey),
	)
	t.Cleanup(tmpDir.Remove)

	return tmpDir
}

func TestCrypto_RSACrypt(t *testing.T) {
	tmpDir := setupCryptoTest(t)
	o, e, err := cmd(t,
		"-i", `{{ crypto.RSAGenerateKey 2048 -}}`,
		"-o", `key.pem`).
		withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "")

	o, e, err = cmd(t,
		"-c", "privKey=./key.pem?type=text/plain",
		"-i", `{{ $pub := crypto.RSADerivePublicKey .privKey -}}
{{ $enc := "hello" | crypto.RSAEncrypt $pub -}}
{{ crypto.RSADecryptBytes .privKey $enc | conv.ToString }}
{{ crypto.RSADecrypt .privKey $enc }}
`).
		withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "hello\nhello\n")
}

func TestCrypto_DerivePublicKey(t *testing.T) {
	// Test unified DerivePublicKey with RSA
	o, e, err := cmd(t,
		"-i", `{{ $key := crypto.RSAGenerateKey 2048 -}}
{{ $pub := crypto.DerivePublicKey $key -}}
{{ $pub | strings.HasPrefix "-----BEGIN PUBLIC KEY-----" }}`).run()
	assertSuccess(t, o, e, err, "true")

	// Test unified DerivePublicKey with ECDSA
	o, e, err = cmd(t,
		"-i", `{{ $key := crypto.ECDSAGenerateKey -}}
{{ $pub := crypto.DerivePublicKey $key -}}
{{ $pub | strings.HasPrefix "-----BEGIN PUBLIC KEY-----" }}`).run()
	assertSuccess(t, o, e, err, "true")

	// Test unified DerivePublicKey with Ed25519
	o, e, err = cmd(t,
		"-i", `{{ $key := crypto.Ed25519GenerateKey -}}
{{ $pub := crypto.DerivePublicKey $key -}}
{{ $pub | strings.HasPrefix "-----BEGIN PUBLIC KEY-----" }}`).run()
	assertSuccess(t, o, e, err, "true")
}
