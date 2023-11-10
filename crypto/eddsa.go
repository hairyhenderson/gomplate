package crypto

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// Ed25519GenerateKey returns a random PEM encoded Ed25519 Private Key.
func Ed25519GenerateKey() ([]byte, error) {
	_, secret, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("Ed25519GenerateKey: failed to generate EDDSA private key: %w", err)
	}
	return pemEncodeEdPrivateKey(secret)
}

// Ed25519GenerateKeyFromSeed returns a PEM encoded Ed25519 Private Key from
// `seed`. Panics if len(seed) is not SeedSize.
func Ed25519GenerateKeyFromSeed(seed []byte) ([]byte, error) {
	return pemEncodeEdPrivateKey(ed25519.NewKeyFromSeed(seed)) // Panics
}

// Ed25519DerivePublicKey returns an ed25519 Public Key from given PEM encoded
// `privatekey`.
func Ed25519DerivePublicKey(privatekey []byte) ([]byte, error) {
	secret, err := ed25519DecodeFromPEM(privatekey)
	if err != nil {
		return nil, fmt.Errorf("EDDSADerivePublicKey: could not decode private key")
	}
	b, err := x509.MarshalPKIXPublicKey(secret.Public())
	if err != nil {
		return nil, fmt.Errorf("EDDSADerivePublicKey: failed to marshal PKIX public key: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}), nil
}

// pemEncodeEdPrivateKey is a convenience function for PEM encoding `secret`.
func pemEncodeEdPrivateKey(secret ed25519.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(secret)
	if err != nil {
		return nil, fmt.Errorf("pemEncodeEdPrivateKey: failed to marshal ECDSA private key: %w", err)
	}
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	}
	buf := &bytes.Buffer{}
	err = pem.Encode(buf, block)
	if err != nil {
		return nil, fmt.Errorf("pemEncodeEdPrivateKey: failed to encode generated EDDSA private key: pem encoding failed: %w", err)
	}
	return buf.Bytes(), nil
}

// ed25519DecodeFromPEM returns an ed25519.PrivateKey from given PEM encoded
// `privatekey`.
func ed25519DecodeFromPEM(privatekey []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(privatekey)
	if block == nil {
		return nil, fmt.Errorf("ed25519DecodeFromPEM: failed to read key: no key found")
	}
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("ed25519DecodeFromPEM: invalid private key: %w", err)
	}
	secret, ok := priv.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("ed25519DecodeFromPEM: invalid private key: %w", err)
	}
	return secret, nil
}
