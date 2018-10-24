package aws

import (
	"encoding/base64"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/pkg/errors"
)

// KMS -
type KMS struct {
	Client *kms.KMS
}

// NewKMS -
func NewKMS(option ClientOptions) *KMS {
	//return
}

// Encrypt plaintext using the specified key
func (k *KMS) Encrypt(keyID string, plaintext string) (string, error) {

}

// Decrypt cyphertext
func (k *KMS) Decrypt(cyphertext string) (string, error) {

}
