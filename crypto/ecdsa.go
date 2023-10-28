package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// Curves -
func Curves(c string) (elliptic.Curve, bool) {
	switch c {
	case "P224":
		return elliptic.P224(), true
	case "P256":
		return elliptic.P256(), true
	case "P384":
		return elliptic.P384(), true
	case "P521":
		return elliptic.P521(), true
	default:
		return nil, false
	}
}

// ECDSAGenerateKey -
func ECDSAGenerateKey(curve elliptic.Curve) ([]byte, error) {
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDSA private key: %w", err)
	}

	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ECDSA private key: %w", err)
	}

	block := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	}
	buf := &bytes.Buffer{}

	err = pem.Encode(buf, block)
	if err != nil {
		return nil, fmt.Errorf("failed to encode generated ECDSA private key: pem encoding failed: %w", err)
	}

	return buf.Bytes(), nil
}

// ECDSADerivePublicKey -
func ECDSADerivePublicKey(privatekey []byte) ([]byte, error) {
	block, _ := pem.Decode(privatekey)
	if block == nil {
		return nil, fmt.Errorf("failed to read key: no key found")
	}

	priv, err := x509.ParseECPrivateKey(block.Bytes)
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
