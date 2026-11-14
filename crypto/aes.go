package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// EncryptAESCBC - use a 128, 192, or 256 bit key to encrypt the given content
// using AES-CBC. The output will not be encoded. Usually the output would be
// base64-encoded for display. Empty content will not be encrypted.
//
// This function is compatible with Helm's decryptAES function, when the output
// is base64-encoded, and when the key is 256 bits long.
func EncryptAESCBC(key []byte, in []byte) ([]byte, error) {
	if len(in) == 0 {
		return in, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Pad the content to be a multiple of the block size, with the pad length
	// as the content. Note that the padding will be a full block when the
	// content is already a multiple of the block size.
	// This algorithm is described in the TLS spec:
	// https://datatracker.ietf.org/doc/html/rfc5246#section-6.2.3.2
	bs := block.BlockSize()
	pl := bs - len(in)%bs

	// Copy the content into the output slice after the IV, and pad it there
	// with pl, repeated pl times. Padding in a copy means the caller's slice
	// is never written to, even if it has spare capacity.
	out := make([]byte, bs+len(in)+pl)
	copy(out[bs:], in)
	for i := bs + len(in); i < len(out); i++ {
		//nolint:gosec // G115 — pl is PKCS#7 length in [1, bs] and bs is the cipher block size
		out[i] = byte(pl)
	}

	// Generate a random IV. Must be the same length as the block size, and is
	// stored at the beginning of the output slice unencrypted, so that it can
	// be used for decryption.
	iv := out[:bs]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// encrypt the padded content in place
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(out[bs:], out[bs:])

	return out, nil
}

// DecryptAESCBC - use a 128, 192, or 256 bit key to decrypt the given content
// using AES-CBC. The output will not be encoded. Empty content will not be
// decrypted.
//
// An error is returned if the content can't be the output of EncryptAESCBC,
// i.e. if it isn't a whole number of blocks, is too short to hold an IV and at
// least one block of ciphertext, or doesn't end with valid padding (which is
// what usually happens when the wrong key is used).
//
// This function is compatible with Helm's encryptAES function, when the input
// is base64-decoded, and when the key is 256 bits long.
func DecryptAESCBC(key []byte, in []byte) ([]byte, error) {
	if len(in) == 0 {
		return nil, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bs := block.BlockSize()
	if len(in) < 2*bs || len(in)%bs != 0 {
		return nil, fmt.Errorf("invalid ciphertext length %d: must be a multiple of %d bytes, and hold an IV and at least one block", len(in), bs)
	}

	// the first block is the IV, unencrypted
	iv := in[:bs]

	// the rest of the content is encrypted
	in = in[bs:]

	out := make([]byte, len(in))

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(out, in)

	// content must always be padded with at least one byte, and the padding
	// byte must be the padding length
	pl := int(out[len(out)-1])
	if pl == 0 || pl > bs {
		return nil, errors.New("invalid padding")
	}

	for _, b := range out[len(out)-pl:] {
		if int(b) != pl {
			return nil, errors.New("invalid padding")
		}
	}

	return out[:len(out)-pl], nil
}
