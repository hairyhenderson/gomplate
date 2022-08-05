package vault

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hairyhenderson/gomplate/v3/aws"
	"github.com/hairyhenderson/gomplate/v3/conv"
	"github.com/hairyhenderson/gomplate/v3/env"
	"github.com/hairyhenderson/gomplate/v3/internal/iohelpers"
	"github.com/pkg/errors"
)

// GetToken -
func (v *Vault) GetToken() (string, error) {
	// sorted in order of precedence
	authFuncs := []func() (string, error){
		v.AppRoleLogin,
		v.AppIDLogin,
		v.GitHubLogin,
		v.UserPassLogin,
		v.TokenLogin,
		v.EC2Login,
	}
	for _, f := range authFuncs {
		if token, err := f(); token != "" || err != nil {
			return token, err
		}
	}
	return "", errors.New("no vault auth methods succeeded")
}

// AppIDLogin - app-id auth backend
func (v *Vault) AppIDLogin() (string, error) {
	appID := env.Getenv("VAULT_APP_ID")
	userID := env.Getenv("VAULT_USER_ID")

	if appID == "" || userID == "" {
		return "", nil
	}

	mount := env.Getenv("VAULT_AUTH_APP_ID_MOUNT", "app-id")

	vars := map[string]interface{}{
		"user_id": userID,
	}

	path := fmt.Sprintf("auth/%s/login/%s", mount, appID)
	secret, err := v.client.Logical().Write(path, vars)
	if err != nil {
		return "", errors.Wrapf(err, "appID logon failed")
	}
	if secret == nil {
		return "", errors.New("empty response from AppID logon")
	}

	return secret.Auth.ClientToken, nil
}

// AppRoleLogin - approle auth backend
func (v *Vault) AppRoleLogin() (string, error) {
	roleID := env.Getenv("VAULT_ROLE_ID")
	secretID := env.Getenv("VAULT_SECRET_ID")

	if roleID == "" || secretID == "" {
		return "", nil
	}

	mount := env.Getenv("VAULT_AUTH_APPROLE_MOUNT", "approle")

	vars := map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	path := fmt.Sprintf("auth/%s/login", mount)
	secret, err := v.client.Logical().Write(path, vars)
	if err != nil {
		return "", errors.Wrap(err, "appRole logon failed")
	}
	if secret == nil {
		return "", errors.New("empty response from AppRole logon")
	}

	return secret.Auth.ClientToken, nil
}

// GitHubLogin - github auth backend
func (v *Vault) GitHubLogin() (string, error) {
	githubToken := env.Getenv("VAULT_AUTH_GITHUB_TOKEN")

	if githubToken == "" {
		return "", nil
	}

	mount := env.Getenv("VAULT_AUTH_GITHUB_MOUNT", "github")

	vars := map[string]interface{}{
		"token": githubToken,
	}

	path := fmt.Sprintf("auth/%s/login", mount)
	secret, err := v.client.Logical().Write(path, vars)
	if err != nil {
		return "", errors.Wrap(err, "appRole logon failed")
	}
	if secret == nil {
		return "", errors.New("empty response from AppRole logon")
	}

	return secret.Auth.ClientToken, nil
}

// UserPassLogin - userpass auth backend
func (v *Vault) UserPassLogin() (string, error) {
	username := env.Getenv("VAULT_AUTH_USERNAME")
	password := env.Getenv("VAULT_AUTH_PASSWORD")

	if username == "" || password == "" {
		return "", nil
	}

	mount := env.Getenv("VAULT_AUTH_USERPASS_MOUNT", "userpass")

	vars := map[string]interface{}{
		"password": password,
	}

	path := fmt.Sprintf("auth/%s/login/%s", mount, username)
	secret, err := v.client.Logical().Write(path, vars)
	if err != nil {
		return "", errors.Wrap(err, "userPass logon failed")
	}
	if secret == nil {
		return "", errors.New("empty response from UserPass logon")
	}

	return secret.Auth.ClientToken, nil
}

// EC2Login - AWS EC2 auth backend
func (v *Vault) EC2Login() (string, error) {
	mount := env.Getenv("VAULT_AUTH_AWS_MOUNT", "aws")
	output := env.Getenv("VAULT_AUTH_AWS_NONCE_OUTPUT")

	nonce := env.Getenv("VAULT_AUTH_AWS_NONCE")

	vars, err := createEc2LoginVars(nonce)
	if err != nil {
		return "", err
	}
	if vars["pkcs7"] == "" {
		return "", nil
	}

	path := fmt.Sprintf("auth/%s/login", mount)
	secret, err := v.client.Logical().Write(path, vars)
	if err != nil {
		return "", errors.Wrapf(err, "AWS EC2 logon failed")
	}
	if secret == nil {
		return "", errors.New("empty response from AWS EC2 logon")
	}

	if output != "" {
		if val, ok := secret.Auth.Metadata["nonce"]; ok {
			nonce = val
		}
		f, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, iohelpers.NormalizeFileMode(0o600))
		if err != nil {
			return "", errors.Wrapf(err, "Error opening nonce output file")
		}
		n, err := f.Write([]byte(nonce + "\n"))
		if err != nil {
			return "", errors.Wrapf(err, "Error writing nonce output file")
		}
		if n == 0 {
			return "", errors.Wrapf(err, "No bytes written to nonce output file")
		}
	}

	return secret.Auth.ClientToken, nil
}

func createEc2LoginVars(nonce string) (map[string]interface{}, error) {
	role := env.Getenv("VAULT_AUTH_AWS_ROLE")

	vars := map[string]interface{}{}

	if role != "" {
		vars["role"] = role
	}

	if nonce != "" {
		vars["nonce"] = nonce
	}

	opts := aws.ClientOptions{
		Timeout: time.Duration(conv.MustAtoi(env.Getenv("AWS_TIMEOUT"))) * time.Millisecond,
	}

	meta := aws.NewEc2Meta(opts)

	doc, err := meta.Dynamic("instance-identity/pkcs7")
	if err != nil {
		return nil, err
	}
	vars["pkcs7"] = strings.ReplaceAll(strings.TrimSpace(doc), "\n", "")
	return vars, nil
}

// TokenLogin -
func (v *Vault) TokenLogin() (string, error) {
	if token := env.Getenv("VAULT_TOKEN"); token != "" {
		return token, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	f, err := os.OpenFile(path.Join(homeDir, ".vault-token"), os.O_RDONLY, 0)
	if err != nil {
		return "", nil
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
