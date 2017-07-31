package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/blang/vfs"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/helper/jsonutil"
)

// logFatal is defined so log.Fatal calls can be overridden for testing
var logFatal = log.Fatal

// Vault -
type Vault struct {
	client *vaultapi.Client
	fs     vfs.Filesystem
}

// NewVault - instantiate a new
func NewVault() *Vault {
	vaultConfig := vaultapi.DefaultConfig()

	err := vaultConfig.ReadEnvironment()
	if err != nil {
		logFatal("Vault setup failed", err)
	}

	client, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		logFatal("Vault setup failed", err)
	}

	return &Vault{client, nil}
}

// Login -
func (v *Vault) Login() {
	v.client.SetToken(v.GetToken())
}

// GetToken -
func (v *Vault) GetToken(fsOverrides ...vfs.Filesystem) string {
	if len(fsOverrides) == 0 {
		v.fs = vfs.OS()
	} else {
		v.fs = fsOverrides[0]
	}

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
	logFatal("All vault auth failed")
	return ""
}

// Logout -
func (v *Vault) Logout() {
}

func (v *Vault) Read(path string) ([]byte, error) {
	secret, err := v.client.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	return jsonutil.EncodeJSON(secret.Data)
}

// AppIDLogin - app-id auth backend
func (v *Vault) AppIDLogin() string {
	env := &Env{}
	appID := env.GetenvFile(v.fs, "VAULT_APP_ID", "")
	userID := env.GetenvFile(v.fs, "VAULT_USER_ID", "")

	if appID == "" {
		return ""
	}
	if userID == "" {
		return ""
	}

	mount := env.GetenvFile(v.fs, "VAULT_AUTH_APP_ID_MOUNT", "app-id")

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
	env := &Env{}
	roleID := env.GetenvFile(v.fs, "VAULT_ROLE_ID", "")
	secretID := env.GetenvFile(v.fs, "VAULT_SECRET_ID", "")

	if roleID == "" {
		return ""
	}
	if secretID == "" {
		return ""
	}

	mount := env.GetenvFile(v.fs, "VAULT_AUTH_APPROLE_MOUNT", "approle")

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
	env := &Env{}
	githubToken := env.GetenvFile(v.fs, "VAULT_AUTH_GITHUB_TOKEN", "")

	if githubToken == "" {
		return ""
	}

	mount := env.GetenvFile(v.fs, "VAULT_AUTH_GITHUB_MOUNT", "github")

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
	env := &Env{}
	username := env.GetenvFile(v.fs, "VAULT_AUTH_USERNAME", "")
	password := env.GetenvFile(v.fs, "VAULT_AUTH_PASSWORD", "")

	if username == "" {
		return ""
	}
	if password == "" {
		return ""
	}

	mount := env.GetenvFile(v.fs, "VAULT_AUTH_USERPASS_MOUNT", "userpass")

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

// TokenLogin -
func (v *Vault) TokenLogin() string {
	env := &Env{}
	if token := env.GetenvFile(v.fs, "VAULT_TOKEN", ""); token != "" {
		return token
	}
	f, err := v.fs.OpenFile(path.Join(v.homeDir(), ".vault-token"), os.O_RDONLY, 0)
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
