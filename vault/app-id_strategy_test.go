package vault

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppIDAuthStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_APP_ID")
	defer os.Unsetenv("VAULT_USER_ID")
	defer os.Unsetenv("VAULT_AUTH_APP_ID_MOUNT")

	os.Unsetenv("VAULT_APP_ID")
	os.Unsetenv("VAULT_USER_ID")
	assert.Nil(t, NewAppIDAuthStrategy())

	os.Setenv("VAULT_APP_ID", "foo")
	assert.Nil(t, NewAppIDAuthStrategy())

	os.Unsetenv("VAULT_APP_ID")
	os.Setenv("VAULT_USER_ID", "bar")
	assert.Nil(t, NewAppIDAuthStrategy())

	os.Setenv("VAULT_APP_ID", "foo")
	os.Setenv("VAULT_USER_ID", "bar")
	auth := NewAppIDAuthStrategy()
	assert.Equal(t, "foo", auth.AppID)
	assert.Equal(t, "bar", auth.UserID)
	assert.Equal(t, "app-id", auth.Mount)

	os.Setenv("VAULT_APP_ID", "baz")
	os.Setenv("VAULT_USER_ID", "qux")
	os.Setenv("VAULT_AUTH_APP_ID_MOUNT", "quux")
	auth = NewAppIDAuthStrategy()
	assert.Equal(t, "baz", auth.AppID)
	assert.Equal(t, "qux", auth.UserID)
	assert.Equal(t, "quux", auth.Mount)
}
