package crypto

import (
	"crypto"
	"crypto/pbkdf2"
	"crypto/sha1" //nolint: gosec
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"sync"
)

var hashFuncs = sync.OnceValue(func() map[crypto.Hash]func() hash.Hash {
	h := make(map[crypto.Hash]func() hash.Hash)
	h[crypto.SHA1] = sha1.New
	h[crypto.SHA224] = sha256.New224
	h[crypto.SHA256] = sha256.New
	h[crypto.SHA384] = sha512.New384
	h[crypto.SHA512] = sha512.New
	h[crypto.SHA512_224] = sha512.New512_224
	h[crypto.SHA512_256] = sha512.New512_256

	return h
})()

// StrToHash - find a hash given a certain string
func StrToHash(hash string) (crypto.Hash, error) {
	switch hash {
	case "SHA1", "SHA-1":
		return crypto.SHA1, nil
	case "SHA224", "SHA-224":
		return crypto.SHA224, nil
	case "SHA256", "SHA-256":
		return crypto.SHA256, nil
	case "SHA384", "SHA-384":
		return crypto.SHA384, nil
	case "SHA512", "SHA-512":
		return crypto.SHA512, nil
	case "SHA512_224", "SHA512/224", "SHA-512_224", "SHA-512/224":
		return crypto.SHA512_224, nil
	case "SHA512_256", "SHA512/256", "SHA-512_256", "SHA-512/256":
		return crypto.SHA512_256, nil
	}
	return 0, fmt.Errorf("no such hash %s", hash)
}

// PBKDF2 - Run the Password-Based Key Derivation Function #2 as defined in
// RFC 8018 (PKCS #5 v2.1)
func PBKDF2(password, salt []byte, iter, keylen int, hashFunc crypto.Hash) ([]byte, error) {
	h, ok := hashFuncs[hashFunc]
	if !ok {
		return nil, fmt.Errorf("hashFunc not supported: %v", hashFunc)
	}

	return pbkdf2.Key(h, string(password), salt, iter, keylen)
}
