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
		return nil, fmt.Errorf("generateKey: %w", err)
	}
	return pemEncodeEdPrivateKey(secret)
}

// Ed25519GenerateKeyFromSeed returns a PEM encoded Ed25519 Private Key from
// `seed`. Returns error if len(seed) is not ed25519.SeedSize (32).
func Ed25519GenerateKeyFromSeed(seed []byte) ([]byte, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("generateKeyFromSeed: incorrect seed size - given: %d wanted %d", len(seed), ed25519.SeedSize)
	}
	return pemEncodeEdPrivateKey(ed25519.NewKeyFromSeed(seed))
}

// Ed25519DerivePublicKey returns an ed25519 Public Key from given PEM encoded
// `privatekey`.
func Ed25519DerivePublicKey(privatekey []byte) ([]byte, error) {
	secret, err := ed25519DecodeFromPEM(privatekey)
	if err != nil {
		return nil, fmt.Errorf("ed25519DecodeFromPEM: could not decode private key: %w", err)
	}
	b, err := x509.MarshalPKIXPublicKey(secret.Public())
	if err != nil {
		return nil, fmt.Errorf("marshalPKIXPublicKey: failed to marshal PKIX public key: %w", err)
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
		return nil, fmt.Errorf("marshalPKCS8PrivateKey: failed to marshal ed25519 private key: %w", err)
	}
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	}
	buf := &bytes.Buffer{}
	err = pem.Encode(buf, block)
	if err != nil {
		return nil, fmt.Errorf("encode: PEM encoding: %w", err)
	}
	return buf.Bytes(), nil
}

// ed25519DecodeFromPEM returns an ed25519.PrivateKey from given PEM encoded
// `privatekey`.
func ed25519DecodeFromPEM(privatekey []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(privatekey)
	if block == nil {
		return nil, fmt.Errorf("decode: failed to read key")
	}
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsePKCS8PrivateKey: invalid private key: %w", err)
	}
	secret, ok := priv.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("ed25519DecodeFromPEM: invalid ed25519 Private Key - given type: %T", priv)
	}
	return secret, nil
}
