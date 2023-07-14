package funcs

import (
	"context"
	"crypto/sha1" //nolint: gosec
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
)

// CryptoNS - the crypto namespace
// Deprecated: don't use
func CryptoNS() *CryptoFuncs {
	return &CryptoFuncs{}
}

// AddCryptoFuncs -
// Deprecated: use CreateCryptoFuncs instead
func AddCryptoFuncs(f map[string]interface{}) {
	for k, v := range CreateCryptoFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateCryptoFuncs -
func CreateCryptoFuncs(ctx context.Context) map[string]interface{} {
	f := map[string]interface{}{}

	ns := &CryptoFuncs{ctx}

	f["crypto"] = func() interface{} { return ns }
	return f
}

// CryptoFuncs -
type CryptoFuncs struct {
	ctx context.Context
}

// SHA1 - Note: SHA-1 is cryptographically broken and should not be used for secure applications.
func (f CryptoFuncs) SHA1(input interface{}) string {
	// nolint: gosec
	out, _ := f.SHA1Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA224 -
func (f CryptoFuncs) SHA224(input interface{}) string {
	out, _ := f.SHA224Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA256 -
func (f CryptoFuncs) SHA256(input interface{}) string {
	out, _ := f.SHA256Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA384 -
func (f CryptoFuncs) SHA384(input interface{}) string {
	out, _ := f.SHA384Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA512 -
func (f CryptoFuncs) SHA512(input interface{}) string {
	out, _ := f.SHA512Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA512_224 -
// nolint: revive,stylecheck
func (f CryptoFuncs) SHA512_224(input interface{}) string {
	out, _ := f.SHA512_224Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA512_256 -
// nolint: revive,stylecheck
func (f CryptoFuncs) SHA512_256(input interface{}) string {
	out, _ := f.SHA512_256Bytes(input)
	return fmt.Sprintf("%02x", out)
}

// SHA1 - Note: SHA-1 is cryptographically broken and should not be used for secure applications.
func (CryptoFuncs) SHA1Bytes(input interface{}) ([]byte, error) {
	//nolint:gosec
	b := sha1.Sum(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}

// SHA224 -
func (CryptoFuncs) SHA224Bytes(input interface{}) ([]byte, error) {
	b := sha256.Sum224(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}

// SHA256 -
func (CryptoFuncs) SHA256Bytes(input interface{}) ([]byte, error) {
	b := sha256.Sum256(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}

// SHA384 -
func (CryptoFuncs) SHA384Bytes(input interface{}) ([]byte, error) {
	b := sha512.Sum384(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}

// SHA512 -
func (CryptoFuncs) SHA512Bytes(input interface{}) ([]byte, error) {
	b := sha512.Sum512(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}

// SHA512_224 -
// nolint: revive,stylecheck
func (CryptoFuncs) SHA512_224Bytes(input interface{}) ([]byte, error) {
	b := sha512.Sum512_224(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}

// SHA512_256 -
// nolint: revive,stylecheck
func (CryptoFuncs) SHA512_256Bytes(input interface{}) ([]byte, error) {
	b := sha512.Sum512_256(toBytes(input))
	out := make([]byte, len(b))
	copy(out, b[:])
	return out, nil
}
