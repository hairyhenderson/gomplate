package vault

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserPassAuthStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_AUTH_USERNAME")
	defer os.Unsetenv("VAULT_AUTH_PASSWORD")
	defer os.Unsetenv("VAULT_AUTH_USERPASS_MOUNT")

	assert.Nil(t, NewUserPassAuthStrategy())

	os.Setenv("VAULT_AUTH_USERNAME", "foo")
	assert.Nil(t, NewUserPassAuthStrategy())

	os.Unsetenv("VAULT_AUTH_USERNAME")
	os.Setenv("VAULT_AUTH_PASSWORD", "bar")
	assert.Nil(t, NewUserPassAuthStrategy())

	os.Setenv("VAULT_AUTH_USERNAME", "foo")
	os.Setenv("VAULT_AUTH_PASSWORD", "bar")
	auth := NewUserPassAuthStrategy()
	assert.Equal(t, "foo", auth.Username)
	assert.Equal(t, "bar", auth.Password)
	assert.Equal(t, "userpass", auth.Mount)

	os.Setenv("VAULT_AUTH_USERNAME", "foo")
	os.Setenv("VAULT_AUTH_PASSWORD", "bar")
	os.Setenv("VAULT_AUTH_USERPASS_MOUNT", "baz")
	auth = NewUserPassAuthStrategy()
	assert.Equal(t, "foo", auth.Username)
	assert.Equal(t, "bar", auth.Password)
	assert.Equal(t, "baz", auth.Mount)
}
