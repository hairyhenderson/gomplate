package vault

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppRoleAuthStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_ROLE_ID")
	defer os.Unsetenv("VAULT_SECRET_ID")
	defer os.Unsetenv("VAULT_AUTH_APPROLE_MOUNT")

	os.Unsetenv("VAULT_ROLE_ID")
	os.Unsetenv("VAULT_SECRET_ID")
	assert.Nil(t, NewAppRoleAuthStrategy())

	os.Setenv("VAULT_ROLE_ID", "foo")
	assert.Nil(t, NewAppRoleAuthStrategy())

	os.Unsetenv("VAULT_ROLE_ID")
	os.Setenv("VAULT_SECRET_ID", "bar")
	assert.Nil(t, NewAppRoleAuthStrategy())

	os.Setenv("VAULT_ROLE_ID", "foo")
	os.Setenv("VAULT_SECRET_ID", "bar")
	auth := NewAppRoleAuthStrategy()
	assert.Equal(t, "foo", auth.RoleID)
	assert.Equal(t, "bar", auth.SecretID)
	assert.Equal(t, "approle", auth.Mount)

	os.Setenv("VAULT_ROLE_ID", "baz")
	os.Setenv("VAULT_SECRET_ID", "qux")
	os.Setenv("VAULT_AUTH_APPROLE_MOUNT", "quux")
	auth = NewAppRoleAuthStrategy()
	assert.Equal(t, "baz", auth.RoleID)
	assert.Equal(t, "qux", auth.SecretID)
	assert.Equal(t, "quux", auth.Mount)
}
