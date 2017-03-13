package vault

import (
	"fmt"
	"os"
)

// UserPassAuthStrategy - an AuthStrategy that uses Vault's userpass authentication backend.
type UserPassAuthStrategy struct {
	*Strategy
	Username string `json:"-"`
	Password string `json:"password"`
}

// NewUserPassAuthStrategy - create an AuthStrategy that uses Vault's userpass auth
// backend.
func NewUserPassAuthStrategy() *UserPassAuthStrategy {
	username := os.Getenv("VAULT_AUTH_USERNAME")
	password := os.Getenv("VAULT_AUTH_PASSWORD")
	mount := os.Getenv("VAULT_AUTH_USERPASS_MOUNT")
	if mount == "" {
		mount = "userpass"
	}
	if username != "" && password != "" {
		return &UserPassAuthStrategy{&Strategy{mount, nil}, username, password}
	}
	return nil
}

func (a *UserPassAuthStrategy) String() string {
	return fmt.Sprintf("username: %s, password: %s, mount: %s", a.Username, a.Password, a.Mount)
}
