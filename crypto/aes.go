package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
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

	// pad with pl, repeated pl times
	in = append(in, bytes.Repeat([]byte{byte(pl)}, pl)...)

	out := make([]byte, bs+len(in))

	// Generate a random IV. Must be the same length as the block size, and is
	// stored at the beginning of the output slice unencrypted, so that it can
	// be used for decryption.
	iv := out[:bs]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// encrypt the content into the rest of the output slice
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(out[bs:], in)

	return out, nil
}

// DecryptAESCBC - use a 128, 192, or 256 bit key to decrypt the given content
// using AES-CBC. The output will not be encoded. Empty content will not be
// decrypted.
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

	// the first block is the IV, unencrypted
	iv := in[:aes.BlockSize]

	// the rest of the content is encrypted
	in = in[aes.BlockSize:]

	out := make([]byte, len(in))

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(out, in)

	// content must always be padded with at least one byte, and the padding
	// byte must be the padding length
	pl := int(out[len(out)-1])
	out = out[:len(out)-pl]

	return out, nil
}
