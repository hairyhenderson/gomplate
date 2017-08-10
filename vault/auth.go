package vault

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/blang/vfs"
	"github.com/hairyhenderson/gomplate/aws"
	"github.com/hairyhenderson/gomplate/conv"
	"github.com/hairyhenderson/gomplate/env"
)

// GetToken -
func (v *Vault) GetToken() string {
	if token := v.AppRoleLogin(); token != "" {
		return token
	}
	if token := v.AppIDLogin(); token != "" {
		return token
	}
	if token := v.GitHubLogin(); token != "" {
		return token
	}
	if token := v.UserPassLogin(); token != "" {
		return token
	}
	if token := v.TokenLogin(); token != "" {
		return token
	}
	if token := v.EC2Login(); token != "" {
		return token
	}
	logFatal("All vault auth failed")
	return ""
}

// AppIDLogin - app-id auth backend
func (v *Vault) AppIDLogin() string {
	appID := env.Getenv("VAULT_APP_ID")
	userID := env.Getenv("VAULT_USER_ID")

	if appID == "" {
		return ""
	}
	if userID == "" {
		return ""
	}

	mount := env.Getenv("VAULT_AUTH_APP_ID_MOUNT", "app-id")

	vars := map[string]interface{}{
		"user_id": userID,
	}

	path := fmt.Sprintf("auth/%s/login/%s", mount, appID)
	secret, err := v.client.Logical().Write(path, vars)
	if err != nil {
		logFatal("AppID logon failed", err)
	}
	if secret == nil {
		logFatal("Empty response from AppID logon")
	}

	return secret.Auth.ClientToken
}

// AppRoleLogin - approle auth backend
func (v *Vault) AppRoleLogin() string {
	roleID := env.Getenv("VAULT_ROLE_ID")
	secretID := env.Getenv("VAULT_SECRET_ID")

	if roleID == "" {
		return ""
	}
	if secretID == "" {
		return ""
	}

	mount := env.Getenv("VAULT_AUTH_APPROLE_MOUNT", "approle")

	vars := map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	path := fmt.Sprintf("auth/%s/login", mount)
	secret, err := v.client.Logical().Write(path, vars)
	if err != nil {
		logFatal("AppRole logon failed", err)
	}
	if secret == nil {
		logFatal("Empty response from AppRole logon")
	}

	return secret.Auth.ClientToken
}

// GitHubLogin - github auth backend
func (v *Vault) GitHubLogin() string {
	githubToken := env.Getenv("VAULT_AUTH_GITHUB_TOKEN")

	if githubToken == "" {
		return ""
	}

	mount := env.Getenv("VAULT_AUTH_GITHUB_MOUNT", "github")

	vars := map[string]interface{}{
		"token": githubToken,
	}

	path := fmt.Sprintf("auth/%s/login", mount)
	secret, err := v.client.Logical().Write(path, vars)
	if err != nil {
		logFatal("AppRole logon failed", err)
	}
	if secret == nil {
		logFatal("Empty response from AppRole logon")
	}

	return secret.Auth.ClientToken
}

// UserPassLogin - userpass auth backend
func (v *Vault) UserPassLogin() string {
	username := env.Getenv("VAULT_AUTH_USERNAME")
	password := env.Getenv("VAULT_AUTH_PASSWORD")

	if username == "" {
		return ""
	}
	if password == "" {
		return ""
	}

	mount := env.Getenv("VAULT_AUTH_USERPASS_MOUNT", "userpass")

	vars := map[string]interface{}{
		"password": password,
	}

	path := fmt.Sprintf("auth/%s/login/%s", mount, username)
	secret, err := v.client.Logical().Write(path, vars)
	if err != nil {
		logFatal("UserPass logon failed", err)
	}
	if secret == nil {
		logFatal("Empty response from UserPass logon")
	}

	return secret.Auth.ClientToken
}

// EC2Login - AWS EC2 auth backend
func (v *Vault) EC2Login() string {
	role := env.Getenv("VAULT_AUTH_AWS_ROLE")
	mount := env.Getenv("VAULT_AUTH_AWS_MOUNT", "aws")

	vars := map[string]interface{}{}

	if role != "" {
		vars["role"] = role
	}

	opts := aws.ClientOptions{
		Timeout: time.Duration(conv.MustAtoi(os.Getenv("AWS_TIMEOUT"))) * time.Millisecond,
	}

	meta := aws.NewEc2Meta(opts)

	vars["pkcs7"] = strings.Replace(strings.TrimSpace(meta.Dynamic("instance-identity/pkcs7")), "\n", "", -1)

	if vars["pkcs7"] == "" {
		return ""
	}

	path := fmt.Sprintf("auth/%s/login", mount)
	secret, err := v.client.Logical().Write(path, vars)
	if err != nil {
		logFatal("AWS EC2 logon failed", err)
	}
	if secret == nil {
		logFatal("Empty response from AWS EC2 logon")
	}

	return secret.Auth.ClientToken
}

// TokenLogin -
func (v *Vault) TokenLogin() string {
	if token := env.Getenv("VAULT_TOKEN"); token != "" {
		return token
	}
	fs := vfs.OS()
	f, err := fs.OpenFile(path.Join(v.homeDir(), ".vault-token"), os.O_RDONLY, 0)
	if err != nil {
		return ""
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return ""
	}
	return string(b)
}

func (v *Vault) homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home
	}
	logFatal(`Neither HOME nor USERPROFILE environment variables are set!
		I can't figure out where the current user's home directory is!`)
	return ""
}
