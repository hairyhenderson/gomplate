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

// ECDSAGenerateKey -
func ECDSAGenerateKey(curve string) ([]byte, error) {
	var c elliptic.Curve

	if curve == "P-224" {
		c = elliptic.P224()
	} else if curve == "P-256" {
		c = elliptic.P256()
	} else if curve == "P-384" {
		c = elliptic.P384()
	} else if curve == "P-521" {
		c = elliptic.P521()
	} else {
		return nil, fmt.Errorf("unknow curve: %s", curve)
	}

	priv, err := ecdsa.GenerateKey(c, rand.Reader)
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
