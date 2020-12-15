package funcs

import (
	"context"
	gcrypto "crypto"
	"crypto/sha1" //nolint: gosec
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"github.com/hairyhenderson/gomplate/v3/conv"
	"github.com/pkg/errors"

	"github.com/hairyhenderson/gomplate/v3/crypto"
)

var (
	cryptoNS     *CryptoFuncs
	cryptoNSInit sync.Once
)

// CryptoNS - the crypto namespace
func CryptoNS() *CryptoFuncs {
	cryptoNSInit.Do(func() { cryptoNS = &CryptoFuncs{} })
	return cryptoNS
}

// AddCryptoFuncs -
func AddCryptoFuncs(f map[string]interface{}) {
	for k, v := range CreateCryptoFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateCryptoFuncs -
func CreateCryptoFuncs(ctx context.Context) map[string]interface{} {
	ns := CryptoNS()
	ns.ctx = ctx

	return map[string]interface{}{"crypto": CryptoNS}
}

// CryptoFuncs -
type CryptoFuncs struct {
	ctx context.Context
}

// PBKDF2 - Run the Password-Based Key Derivation Function #2 as defined in
// RFC 2898 (PKCS #5 v2.0). This function outputs the binary result in hex
// format.
func (f *CryptoFuncs) PBKDF2(password, salt, iter, keylen interface{}, hashFunc ...string) (k string, err error) {
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
	i := conv.ToInt(iter)
	kl := conv.ToInt(keylen)

	dk, err := crypto.PBKDF2(pw, s, i, kl, h)
	return fmt.Sprintf("%02x", dk), err
}

// WPAPSK - Convert an ASCII passphrase to WPA PSK for a given SSID
func (f *CryptoFuncs) WPAPSK(ssid, password interface{}) (string, error) {
	return f.PBKDF2(password, ssid, 4096, 32)
}

// SHA1 - Note: SHA-1 is cryptographically broken and should not be used for secure applications.
func (f *CryptoFuncs) SHA1(input interface{}) string {
	in := toBytes(input)
	// nolint: gosec
	out := sha1.Sum(in)
	return fmt.Sprintf("%02x", out)
}

// SHA224 -
func (f *CryptoFuncs) SHA224(input interface{}) string {
	in := toBytes(input)
	out := sha256.Sum224(in)
	return fmt.Sprintf("%02x", out)
}

// SHA256 -
func (f *CryptoFuncs) SHA256(input interface{}) string {
	in := toBytes(input)
	out := sha256.Sum256(in)
	return fmt.Sprintf("%02x", out)
}

// SHA384 -
func (f *CryptoFuncs) SHA384(input interface{}) string {
	in := toBytes(input)
	out := sha512.Sum384(in)
	return fmt.Sprintf("%02x", out)
}

// SHA512 -
func (f *CryptoFuncs) SHA512(input interface{}) string {
	in := toBytes(input)
	out := sha512.Sum512(in)
	return fmt.Sprintf("%02x", out)
}

// SHA512_224 -
//nolint: golint,stylecheck
func (f *CryptoFuncs) SHA512_224(input interface{}) string {
	in := toBytes(input)
	out := sha512.Sum512_224(in)
	return fmt.Sprintf("%02x", out)
}

// SHA512_256 -
//nolint: golint,stylecheck
func (f *CryptoFuncs) SHA512_256(input interface{}) string {
	in := toBytes(input)
	out := sha512.Sum512_256(in)
	return fmt.Sprintf("%02x", out)
}

// Bcrypt -
func (f *CryptoFuncs) Bcrypt(args ...interface{}) (string, error) {
	input := ""
	cost := bcrypt.DefaultCost
	if len(args) == 0 {
		return "", errors.Errorf("bcrypt requires at least an 'input' value")
	}
	if len(args) == 1 {
		input = conv.ToString(args[0])
	}
	if len(args) == 2 {
		cost = conv.ToInt(args[0])
		input = conv.ToString(args[1])
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input), cost)
	return string(hash), err
}

// RSAEncrypt -
// Experimental!
func (f *CryptoFuncs) RSAEncrypt(key string, in interface{}) ([]byte, error) {
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
func (f *CryptoFuncs) RSAGenerateKey(args ...interface{}) (string, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return "", err
	}
	bits := 4096
	if len(args) == 1 {
		bits = conv.ToInt(args[0])
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
