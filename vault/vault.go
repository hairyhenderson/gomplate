package vault

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/pkg/errors"

	vaultapi "github.com/hashicorp/vault/api"
)

// Vault -
type Vault struct {
	client *vaultapi.Client
}

// New -
func New(u *url.URL) (*Vault, error) {
	vaultConfig := vaultapi.DefaultConfig()

	err := vaultConfig.ReadEnvironment()
	if err != nil {
		return nil, errors.Wrapf(err, "Vault setup failed")
	}

	setVaultURL(vaultConfig, u)

	client, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "Vault setup failed")
	}

	return &Vault{client}, nil
}

func setVaultURL(c *vaultapi.Config, u *url.URL) {
	if u != nil && u.Host != "" {
		scheme := "https"
		if u.Scheme == "vault+http" {
			scheme = "http"
		}
		c.Address = scheme + "://" + u.Host
	}
}

// Login -
func (v *Vault) Login() error {
	token, err := v.GetToken()
	if err != nil {
		return err
	}
	v.client.SetToken(token)
	return nil
}

// Logout -
func (v *Vault) Logout() {
	v.client.ClearToken()
}

// Read - returns the value of a given path. If no value is found at the given
// path, returns empty slice.
func (v *Vault) Read(path string) ([]byte, error) {
	secret, err := v.client.Logical().Read(path)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return []byte{}, nil
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(secret.Data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (v *Vault) Write(path string, data map[string]interface{}) ([]byte, error) {
	secret, err := v.client.Logical().Write(path, data)
	if secret == nil {
		return []byte{}, err
	}
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(secret.Data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// List -
func (v *Vault) List(path string) ([]byte, error) {
	secret, err := v.client.Logical().List(path)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, nil
	}

	keys, ok := secret.Data["keys"]
	if !ok {
		return nil, errors.Errorf("keys param missing from vault list")
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(keys); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
