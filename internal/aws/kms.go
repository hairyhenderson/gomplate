package aws

import (
	"context"

	b64 "github.com/hairyhenderson/gomplate/v4/base64"

	"github.com/aws/aws-sdk-go-v2/service/kms"
)

// KMSAPI is a subset of kmsiface.KMSAPI
type KMSAPI interface {
	Encrypt(context.Context, *kms.EncryptInput, ...func(*kms.Options)) (*kms.EncryptOutput, error)
	Decrypt(context.Context, *kms.DecryptInput, ...func(*kms.Options)) (*kms.DecryptOutput, error)
}

// KMS is an AWS KMS client
type KMS struct {
	Client KMSAPI
}

// NewKMS - Create new AWS KMS client using an SDKSession
func NewKMS(ctx context.Context) *KMS {
	client := kms.NewFromConfig(SDKConfig(ctx))
	return &KMS{
		Client: client,
	}
}

// Encrypt plaintext using the specified key.
// Returns a base64 encoded ciphertext
func (k *KMS) Encrypt(ctx context.Context, keyID, plaintext string) (string, error) {
	input := &kms.EncryptInput{
		KeyId:     &keyID,
		Plaintext: []byte(plaintext),
	}
	output, err := k.Client.Encrypt(ctx, input)
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
func (k *KMS) Decrypt(ctx context.Context, ciphertext string) (string, error) {
	ciphertextBlob, err := b64.Decode(ciphertext)
	if err != nil {
		return "", err
	}
	input := &kms.DecryptInput{
		CiphertextBlob: ciphertextBlob,
	}
	output, err := k.Client.Decrypt(ctx, input)
	if err != nil {
		return "", err
	}
	return string(output.Plaintext), nil
}
