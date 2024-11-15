package datafs

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/hairyhenderson/go-fsimpl/vaultfs/vaultauth"
	"github.com/hairyhenderson/gomplate/v4/internal/deprecated"
	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/aws"
)

// compositeVaultAuthMethod configures the auth method based on environment
// variables. It extends [vaultfs.EnvAuthMethod] by falling back to AWS EC2
// authentication if the other methods fail.
func compositeVaultAuthMethod(envFsys fs.FS) api.AuthMethod {
	return vaultauth.CompositeAuthMethod(
		vaultauth.EnvAuthMethod(),
		envEC2AuthAdapter(envFsys),
		envIAMAuthAdapter(envFsys),
	)
}

// envEC2AuthAdapter builds an AWS EC2 authentication method from environment
// variables, for use only with [compositeVaultAuthMethod]
func envEC2AuthAdapter(envFS fs.FS) api.AuthMethod {
	mountPath := GetenvFsys(envFS, "VAULT_AUTH_AWS_MOUNT", "aws")

	nonce := GetenvFsys(envFS, "VAULT_AUTH_AWS_NONCE")
	role := GetenvFsys(envFS, "VAULT_AUTH_AWS_ROLE")

	// temporary workaround while we wait to deprecate AWS_META_ENDPOINT
	if endpoint := os.Getenv("AWS_META_ENDPOINT"); endpoint != "" {
		deprecated.WarnDeprecated(context.Background(), "Use AWS_EC2_METADATA_SERVICE_ENDPOINT instead of AWS_META_ENDPOINT")
		if os.Getenv("AWS_EC2_METADATA_SERVICE_ENDPOINT") == "" {
			os.Setenv("AWS_EC2_METADATA_SERVICE_ENDPOINT", endpoint)
		}
	}

	awsauth, err := aws.NewAWSAuth(
		aws.WithEC2Auth(),
		aws.WithMountPath(mountPath),
		aws.WithNonce(nonce),
		aws.WithRole(role),
	)
	if err != nil {
		return nil
	}

	output := GetenvFsys(envFS, "VAULT_AUTH_AWS_NONCE_OUTPUT")
	if output == "" {
		return awsauth
	}

	return &ec2AuthNonceWriter{AWSAuth: awsauth, nonce: nonce, output: output}
}

// envIAMAuthAdapter builds an AWS IAM authentication method from environment
// variables, for use only with [compositeVaultAuthMethod]
func envIAMAuthAdapter(envFS fs.FS) api.AuthMethod {
	mountPath := GetenvFsys(envFS, "VAULT_AUTH_AWS_MOUNT", "aws")
	role := GetenvFsys(envFS, "VAULT_AUTH_AWS_ROLE")

	// temporary workaround while we wait to deprecate AWS_META_ENDPOINT
	if endpoint := os.Getenv("AWS_META_ENDPOINT"); endpoint != "" {
		deprecated.WarnDeprecated(context.Background(), "Use AWS_EC2_METADATA_SERVICE_ENDPOINT instead of AWS_META_ENDPOINT")
		if os.Getenv("AWS_EC2_METADATA_SERVICE_ENDPOINT") == "" {
			os.Setenv("AWS_EC2_METADATA_SERVICE_ENDPOINT", endpoint)
		}
	}

	awsauth, err := aws.NewAWSAuth(
		aws.WithIAMAuth(),
		aws.WithMountPath(mountPath),
		aws.WithRole(role),
	)
	if err != nil {
		return nil
	}

	return awsauth
}

// ec2AuthNonceWriter - wraps an AWSAuth, and writes the nonce to the nonce
// output file - only for ec2 auth
type ec2AuthNonceWriter struct {
	*aws.AWSAuth
	nonce  string
	output string
}

func (a *ec2AuthNonceWriter) Login(ctx context.Context, client *api.Client) (*api.Secret, error) {
	secret, err := a.AWSAuth.Login(ctx, client)
	if err != nil {
		return nil, err
	}

	nonce := a.nonce
	if val, ok := secret.Auth.Metadata["nonce"]; ok {
		nonce = val
	}

	err = os.WriteFile(a.output, []byte(nonce+"\n"), iohelpers.NormalizeFileMode(0o600))
	if err != nil {
		return nil, fmt.Errorf("error writing nonce output file: %w", err)
	}

	return secret, nil
}
