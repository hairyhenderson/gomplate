package vault

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppIDStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_APP_ID")
	defer os.Unsetenv("VAULT_USER_ID")
	defer os.Unsetenv("VAULT_AUTH_APP_ID_MOUNT")

	os.Unsetenv("VAULT_APP_ID")
	os.Unsetenv("VAULT_USER_ID")
	assert.Nil(t, AppIDStrategy())

	os.Setenv("VAULT_APP_ID", "foo")
	assert.Nil(t, AppIDStrategy())

	os.Unsetenv("VAULT_APP_ID")
	os.Setenv("VAULT_USER_ID", "bar")
	assert.Nil(t, AppIDStrategy())

	os.Setenv("VAULT_APP_ID", "foo")
	os.Setenv("VAULT_USER_ID", "bar")
	auth := AppIDStrategy()
	assert.Equal(t, "foo", auth.body["app_id"])
	assert.Equal(t, "bar", auth.body["user_id"])
	assert.Equal(t, "app-id", auth.mount)

	os.Setenv("VAULT_APP_ID", "baz")
	os.Setenv("VAULT_USER_ID", "qux")
	os.Setenv("VAULT_AUTH_APP_ID_MOUNT", "quux")
	auth = AppIDStrategy()
	assert.Equal(t, "baz", auth.body["app_id"])
	assert.Equal(t, "qux", auth.body["user_id"])
	assert.Equal(t, "quux", auth.mount)
}

func TestAppRoleStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_ROLE_ID")
	defer os.Unsetenv("VAULT_SECRET_ID")
	defer os.Unsetenv("VAULT_AUTH_APPROLE_MOUNT")

	os.Unsetenv("VAULT_ROLE_ID")
	os.Unsetenv("VAULT_SECRET_ID")
	assert.Nil(t, AppRoleStrategy())

	os.Setenv("VAULT_ROLE_ID", "foo")
	assert.Nil(t, AppRoleStrategy())

	os.Unsetenv("VAULT_ROLE_ID")
	os.Setenv("VAULT_SECRET_ID", "bar")
	assert.Nil(t, AppRoleStrategy())

	os.Setenv("VAULT_ROLE_ID", "foo")
	os.Setenv("VAULT_SECRET_ID", "bar")
	auth := AppRoleStrategy()
	assert.Equal(t, "foo", auth.body["role_id"])
	assert.Equal(t, "bar", auth.body["secret_id"])
	assert.Equal(t, "approle", auth.mount)

	os.Setenv("VAULT_ROLE_ID", "baz")
	os.Setenv("VAULT_SECRET_ID", "qux")
	os.Setenv("VAULT_AUTH_APPROLE_MOUNT", "quux")
	auth = AppRoleStrategy()
	assert.Equal(t, "baz", auth.body["role_id"])
	assert.Equal(t, "qux", auth.body["secret_id"])
	assert.Equal(t, "quux", auth.mount)
}

func TestGitHubStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_AUTH_GITHUB_TOKEN")
	defer os.Unsetenv("VAULT_AUTH_GITHUB_MOUNT")

	os.Unsetenv("VAULT_AUTH_GITHUB_TOKEN")
	assert.Nil(t, GitHubStrategy())

	os.Setenv("VAULT_AUTH_GITHUB_TOKEN", "foo")
	auth := GitHubStrategy()
	assert.Equal(t, "foo", auth.body["token"])
	assert.Equal(t, "github", auth.mount)

	os.Setenv("VAULT_AUTH_GITHUB_MOUNT", "bar")
	auth = GitHubStrategy()
	assert.Equal(t, "foo", auth.body["token"])
	assert.Equal(t, "bar", auth.mount)
}

func TestUserPassStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_AUTH_USERNAME")
	defer os.Unsetenv("VAULT_AUTH_PASSWORD")
	defer os.Unsetenv("VAULT_AUTH_USERPASS_MOUNT")

	assert.Nil(t, UserPassStrategy())

	os.Setenv("VAULT_AUTH_USERNAME", "foo")
	assert.Nil(t, UserPassStrategy())

	os.Unsetenv("VAULT_AUTH_USERNAME")
	os.Setenv("VAULT_AUTH_PASSWORD", "bar")
	assert.Nil(t, UserPassStrategy())

	os.Setenv("VAULT_AUTH_USERNAME", "foo")
	os.Setenv("VAULT_AUTH_PASSWORD", "bar")
	auth := UserPassStrategy()
	assert.Equal(t, "/v1/auth/userpass/login/foo", auth.path)
	assert.Equal(t, "bar", auth.body["password"])
	assert.Equal(t, "userpass", auth.mount)

	os.Setenv("VAULT_AUTH_USERNAME", "foo")
	os.Setenv("VAULT_AUTH_PASSWORD", "bar")
	os.Setenv("VAULT_AUTH_USERPASS_MOUNT", "baz")
	auth = UserPassStrategy()
	assert.Equal(t, "/v1/auth/baz/login/foo", auth.path)
	assert.Equal(t, "bar", auth.body["password"])
	assert.Equal(t, "baz", auth.mount)
}
