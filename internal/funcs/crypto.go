package funcs

import (
	"context"
	gcrypto "crypto"
	"crypto/elliptic"
	"crypto/sha1" //nolint: gosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/hairyhenderson/gomplate/v4/conv"
	"github.com/hairyhenderson/gomplate/v4/crypto"
)

// CreateCryptoFuncs -
func CreateCryptoFuncs(ctx context.Context) map[string]any {
	f := map[string]any{}

	ns := &CryptoFuncs{ctx}

	f["crypto"] = func() any { return ns }
	return f
}

// CryptoFuncs -
type CryptoFuncs struct {
	ctx context.Context
}

// PBKDF2 - Run the Password-Based Key Derivation Function #2 as defined in
// RFC 2898 (PKCS #5 v2.0). This function outputs the binary result in hex
// format.
func (CryptoFuncs) PBKDF2(password, salt, iter, keylen any, hashFunc ...string) (k string, err error) {
	var h gcrypto.Hash
	if len(hashFunc) == 0 {
		h = gcrypto.SHA1
	} else {
		h, err = crypto.StrToHash(hashFunc[0])
		if err != nil {
			return "", err
		}
	}
	pw := toBytes(password)
	s := toBytes(salt)

	i, err := conv.ToInt(iter)
	if err != nil {
		return "", fmt.Errorf("iter must be an integer: %w", err)
	}

	kl, err := conv.ToInt(keylen)
	if err != nil {
		return "", fmt.Errorf("keylen must be an integer: %w", err)
	}

	dk, err := crypto.PBKDF2(pw, s, i, kl, h)
	return fmt.Sprintf("%02x", dk), err
}

// WPAPSK - Convert an ASCII passphrase to WPA PSK for a given SSID
func (f CryptoFuncs) WPAPSK(ssid, password any) (string, error) {
	return f.PBKDF2(password, ssid, 4096, 32)
}

// SHA1 - Note: SHA-1 is cryptographically broken and should not be used for secure applications.
func (f CryptoFuncs) SHA1(input any) string {
	out, _ := f.SHA1Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA224 -
func (f CryptoFuncs) SHA224(input any) string {
	out, _ := f.SHA224Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA256 -
func (f CryptoFuncs) SHA256(input any) string {
	out, _ := f.SHA256Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA384 -
func (f CryptoFuncs) SHA384(input any) string {
	out, _ := f.SHA384Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA512 -
func (f CryptoFuncs) SHA512(input any) string {
	out, _ := f.SHA512Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA512_224 -
//
//nolint:revive
func (f CryptoFuncs) SHA512_224(input any) string {
	out, _ := f.SHA512_224Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA512_256 -
//
//nolint:revive
func (f CryptoFuncs) SHA512_256(input any) string {
	out, _ := f.SHA512_256Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA1 - Note: SHA-1 is cryptographically broken and should not be used for secure applications.
func (CryptoFuncs) SHA1Bytes(input any) ([]byte, error) {
	//nolint:gosec
	b := sha1.Sum(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}

// SHA224 -
func (CryptoFuncs) SHA224Bytes(input any) ([]byte, error) {
	b := sha256.Sum224(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}

// SHA256 -
func (CryptoFuncs) SHA256Bytes(input any) ([]byte, error) {
	b := sha256.Sum256(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}

// SHA384 -
func (CryptoFuncs) SHA384Bytes(input any) ([]byte, error) {
	b := sha512.Sum384(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}

// SHA512 -
func (CryptoFuncs) SHA512Bytes(input any) ([]byte, error) {
	b := sha512.Sum512(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}

// SHA512_224 -
func (CryptoFuncs) SHA512_224Bytes(input any) ([]byte, error) {
	b := sha512.Sum512_224(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}

// SHA512_256 -
func (CryptoFuncs) SHA512_256Bytes(input any) ([]byte, error) {
	b := sha512.Sum512_256(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}

// Bcrypt -
func (CryptoFuncs) Bcrypt(args ...any) (string, error) {
	input := ""

	var err error
	cost := bcrypt.DefaultCost

	switch len(args) {
	case 1:
		input = conv.ToString(args[0])
	case 2:
		cost, err = conv.ToInt(args[0])
		if err != nil {
			return "", fmt.Errorf("bcrypt cost must be an integer: %w", err)
		}

		input = conv.ToString(args[1])
	default:
		return "", fmt.Errorf("wrong number of args: want 1 or 2, got %d", len(args))
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input), cost)
	return string(hash), err
}

// RSAEncrypt -
// Experimental!
func (f *CryptoFuncs) RSAEncrypt(key string, in any) ([]byte, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return nil, err
	}
	msg := toBytes(in)
	return crypto.RSAEncrypt(key, msg)
}

// RSADecrypt -
// Experimental!
func (f *CryptoFuncs) RSADecrypt(key string, in []byte) (string, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return "", err
	}
	out, err := crypto.RSADecrypt(key, in)
	return string(out), err
}

// RSADecryptBytes -
// Experimental!
func (f *CryptoFuncs) RSADecryptBytes(key string, in []byte) ([]byte, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return nil, err
	}
	out, err := crypto.RSADecrypt(key, in)
	return out, err
}

// RSAGenerateKey -
// Experimental!
func (f *CryptoFuncs) RSAGenerateKey(args ...any) (string, error) {
	err := checkExperimental(f.ctx)
	if err != nil {
		return "", err
	}

	bits := 4096
	if len(args) == 1 {
		bits, err = conv.ToInt(args[0])
		if err != nil {
			return "", fmt.Errorf("bits must be an integer: %w", err)
		}
	} else if len(args) > 1 {
		return "", fmt.Errorf("wrong number of args: want 0 or 1, got %d", len(args))
	}

	out, err := crypto.RSAGenerateKey(bits)
	return string(out), err
}

// RSADerivePublicKey -
// Experimental!
func (f *CryptoFuncs) RSADerivePublicKey(privateKey string) (string, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return "", err
	}
	out, err := crypto.RSADerivePublicKey([]byte(privateKey))
	return string(out), err
}

// ECDSAGenerateKey -
// Experimental!
func (f *CryptoFuncs) ECDSAGenerateKey(args ...any) (string, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return "", err
	}

	curve := elliptic.P256()
	if len(args) == 1 {
		c := conv.ToString(args[0])
		c = strings.ToUpper(c)
		c = strings.ReplaceAll(c, "-", "")
		var ok bool
		curve, ok = crypto.Curves(c)
		if !ok {
			return "", fmt.Errorf("unknown curve: %s", c)
		}
	} else if len(args) > 1 {
		return "", fmt.Errorf("wrong number of args: want 0 or 1, got %d", len(args))
	}

	out, err := crypto.ECDSAGenerateKey(curve)
	return string(out), err
}

// ECDSADerivePublicKey -
// Experimental!
func (f *CryptoFuncs) ECDSADerivePublicKey(privateKey string) (string, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return "", err
	}

	out, err := crypto.ECDSADerivePublicKey([]byte(privateKey))
	return string(out), err
}

// Ed25519GenerateKey -
// Experimental!
func (f *CryptoFuncs) Ed25519GenerateKey() (string, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return "", err
	}
	out, err := crypto.Ed25519GenerateKey()
	return string(out), err
}

// Ed25519GenerateKeyFromSeed -
// Experimental!
func (f *CryptoFuncs) Ed25519GenerateKeyFromSeed(encoding, seed string) (string, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return "", err
	}
	if !utf8.ValidString(seed) {
		return "", fmt.Errorf("given seed is not valid UTF-8") // Don't print out seed (private).
	}
	var seedB []byte
	var err error
	switch encoding {
	case "base64":
		seedB, err = base64.StdEncoding.DecodeString(seed)
	case "hex":
		seedB, err = hex.DecodeString(seed)
	default:
		return "", fmt.Errorf("invalid encoding given: %s - only 'hex' or 'base64' are valid options", encoding)
	}
	if err != nil {
		return "", fmt.Errorf("could not decode given seed: %w", err)
	}
	out, err := crypto.Ed25519GenerateKeyFromSeed(seedB)
	return string(out), err
}

// Ed25519DerivePublicKey -
// Experimental!
func (f *CryptoFuncs) Ed25519DerivePublicKey(privateKey string) (string, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return "", err
	}
	out, err := crypto.Ed25519DerivePublicKey([]byte(privateKey))
	return string(out), err
}

// EncryptAES -
// Experimental!
func (f *CryptoFuncs) EncryptAES(key string, args ...any) ([]byte, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return nil, err
	}

	k, msg, err := parseAESArgs(key, args...)
	if err != nil {
		return nil, err
	}

	return crypto.EncryptAESCBC(k, msg)
}

// DecryptAES -
// Experimental!
func (f *CryptoFuncs) DecryptAES(key string, args ...any) (string, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return "", err
	}

	out, err := f.DecryptAESBytes(key, args...)
	return conv.ToString(out), err
}

// DecryptAESBytes -
// Experimental!
func (f *CryptoFuncs) DecryptAESBytes(key string, args ...any) ([]byte, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return nil, err
	}

	k, msg, err := parseAESArgs(key, args...)
	if err != nil {
		return nil, err
	}

	return crypto.DecryptAESCBC(k, msg)
}

func parseAESArgs(key string, args ...any) ([]byte, []byte, error) {
	keyBits := 256 // default to AES-256-CBC

	var msg []byte

	switch len(args) {
	case 1:
		msg = toBytes(args[0])
	case 2:
		var err error
		keyBits, err = conv.ToInt(args[0])
		if err != nil {
			return nil, nil, fmt.Errorf("keyBits must be an integer: %w", err)
		}
		msg = toBytes(args[1])
	default:
		return nil, nil, fmt.Errorf("wrong number of args: want 2 or 3, got %d", len(args))
	}

	k := make([]byte, keyBits/8)
	copy(k, []byte(key))

	return k, msg, nil
}
