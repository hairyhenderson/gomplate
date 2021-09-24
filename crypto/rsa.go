package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
)

// RSAEncrypt - use the given public key to encrypt the given plaintext. The key
// should be a PEM-encoded RSA public key in PKIX, ASN.1 DER form, typically
// beginning with "PUBLIC KEY". PKCS#1 format is also supported as a fallback.
// The output will not be encoded, so consider base64-encoding it for display.
func RSAEncrypt(key string, in []byte) ([]byte, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, fmt.Errorf("failed to read key %q: no key found", key)
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		if strings.Contains(err.Error(), "use ParsePKCS1PublicKey instead") {
			pub, err = x509.ParsePKCS1PublicKey(block.Bytes)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}
	pubKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key in wrong format, was %T", pub)
	}

	out, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, in)
	return out, err
}

// RSADecrypt - decrypt the ciphertext with the given private key. The key
// must be a PEM-encoded RSA private key in PKCS#1, ASN.1 DER form, typically
// beginning with "RSA PRIVATE KEY". The input text must be plain ciphertext,
// not base64-encoded.
func RSADecrypt(key string, in []byte) ([]byte, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, fmt.Errorf("failed to read key %q: no key found", key)
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	out, err := priv.Decrypt(nil, in, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}
	return out, nil
}

// RSAGenerateKey -
func RSAGenerateKey(bits int) ([]byte, error) {
	// Protect against CWE-326: Inadequate Encryption Strength
	// https://cwe.mitre.org/data/definitions/326.html
	if bits < 2048 {
		return nil, fmt.Errorf("RSA key size must be at least 2048 bits")
	}
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA private key: %w", err)
	}
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}
	buf := &bytes.Buffer{}
	err = pem.Encode(buf, block)
	if err != nil {
		return nil, fmt.Errorf("failed to encode generated RSA private key: pem encoding failed: %w", err)
	}
	return buf.Bytes(), nil
}

// RSADerivePublicKey -
func RSADerivePublicKey(privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, fmt.Errorf("failed to read key: no key found")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	b, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PKIX public key: %w", err)
	}

	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}

	return pem.EncodeToMemory(block), nil
}
