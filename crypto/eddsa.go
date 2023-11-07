package crypto

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// EDDSAGenerateKey returns a random PEM encoded Ed25519 Private Key.
func EDDSAGenerateKey() ([]byte, error) {
	_, secret, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate EDDSA private key: %w", err)
	}
	return pemEncodeEdPrivateKey(secret)
}

// EDDSAGenerateKeyFromSeed returns a PEM encoded Ed25519 Private Key from given
// seed. Panics if len(seed) is not SeedSize.
func EDDSAGenerateKeyFromSeed(seed []byte) ([]byte, error) {
	secret := ed25519.NewKeyFromSeed(seed) // Panics
	return pemEncodeEdPrivateKey(secret)
}

// EDDSADerivePublicKey returns an EDDSA Public Key from a PEM encoded EDDSA
// Private key.
func EDDSADerivePublicKey(privatekey []byte) ([]byte, error) {
	secret, err := edDSADecodeFromPEM(privatekey)
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

// edDSADecodeFromPEM .
func edDSADecodeFromPEM(privatekey []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(privatekey)
	if block == nil {
		return nil, fmt.Errorf("edDSADecodeFromPEM: failed to read key: no key found")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("edDSADecodeFromPEM: invalid private key: %w", err)
	}
	secret, ok := priv.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("edDSADecodeFromPEM: invalid private key: %w", err)
	}
	return secret, nil
}
