package aws

import (
	b64 "github.com/hairyhenderson/gomplate/v4/base64"

	"github.com/aws/aws-sdk-go/service/kms"
)

// KMSAPI is a subset of kmsiface.KMSAPI
type KMSAPI interface {
	Encrypt(input *kms.EncryptInput) (*kms.EncryptOutput, error)
	Decrypt(input *kms.DecryptInput) (*kms.DecryptOutput, error)
}

// KMS is an AWS KMS client
type KMS struct {
	Client KMSAPI
}

// NewKMS - Create new AWS KMS client using an SDKSession
func NewKMS(_ ClientOptions) *KMS {
	client := kms.New(SDKSession())
	return &KMS{
		Client: client,
	}
}

// Encrypt plaintext using the specified key.
// Returns a base64 encoded ciphertext
func (k *KMS) Encrypt(keyID, plaintext string) (string, error) {
	input := &kms.EncryptInput{
		KeyId:     &keyID,
		Plaintext: []byte(plaintext),
	}
	output, err := k.Client.Encrypt(input)
	if err != nil {
		return "", err
	}
	ciphertext, err := b64.Encode(output.CiphertextBlob)
	if err != nil {
		return "", err
	}
	return ciphertext, nil
}

// Decrypt a base64 encoded ciphertext
func (k *KMS) Decrypt(ciphertext string) (string, error) {
	ciphertextBlob, err := b64.Decode(ciphertext)
	if err != nil {
		return "", err
	}
	input := &kms.DecryptInput{
		CiphertextBlob: ciphertextBlob,
	}
	output, err := k.Client.Decrypt(input)
	if err != nil {
		return "", err
	}
	return string(output.Plaintext), nil
}
