package crypto

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// DerivePublicKey derives a public key from any supported private key type
// (RSA, ECDSA, or Ed25519). The key type is auto-detected from the PEM block type.
func DerivePublicKey(privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, fmt.Errorf("failed to read key: no PEM data found")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return RSADerivePublicKey(privateKey)
	case "EC PRIVATE KEY":
		return ECDSADerivePublicKey(privateKey)
	case "PRIVATE KEY":
		// PKCS#8 format - need to parse to determine the key type
		priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
		}

		// Extract the public key using crypto.Signer interface
		signer, ok := priv.(crypto.Signer)
		if !ok {
			return nil, fmt.Errorf("private key does not implement crypto.Signer")
		}

		b, err := x509.MarshalPKIXPublicKey(signer.Public())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal public key: %w", err)
		}

		return pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: b,
		}), nil
	default:
		return nil, fmt.Errorf("unsupported key type: %s", block.Type)
	}
}
