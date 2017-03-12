package vault

import (
	"os"
	"path"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/stretchr/testify/assert"
)

func TestGetenv(t *testing.T) {
	defer os.Unsetenv("FOO")
	defer os.Unsetenv("FOO_FILE")

	secretsDir := "/run/secrets"
	secretPath := path.Join(secretsDir, "secret")
	fs := memfs.Create()
	err := vfs.MkdirAll(fs, secretsDir, 0700)
	assert.NoError(t, err)
	f, err := vfs.Create(fs, secretPath)
	assert.NoError(t, err)
	f.Write([]byte("fromfile"))

	os.Unsetenv("FOO")
	os.Unsetenv("FOO_FILE")
	assert.Empty(t, getValue(fs, "FOO", ""))

	assert.Equal(t, "default", getValue(fs, "FOO", "default"))

	os.Setenv("FOO", "fromenv")
	os.Setenv("FOO_FILE", secretPath)
	assert.Equal(t, "fromenv", getValue(fs, "FOO", "default"))

	os.Unsetenv("FOO")
	assert.Equal(t, "fromfile", getValue(fs, "FOO", "default"))

	os.Setenv("BAR_FILE", "bogusfile")
	assert.Equal(t, "bardefault", getValue(fs, "BAR", "bardefault"))
}

func TestAppIDStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_APP_ID")
	defer os.Unsetenv("VAULT_USER_ID")
	defer os.Unsetenv("VAULT_AUTH_APP_ID_MOUNT")
	defer os.Unsetenv("VAULT_APP_ID_FILE")
	defer os.Unsetenv("VAULT_USER_ID_FILE")

	fs := memfs.Create()

	os.Unsetenv("VAULT_APP_ID")
	os.Unsetenv("VAULT_USER_ID")
	assert.Nil(t, AppIDStrategy(fs))

	os.Setenv("VAULT_APP_ID", "foo")
	assert.Nil(t, AppIDStrategy(fs))

	os.Unsetenv("VAULT_APP_ID")
	os.Setenv("VAULT_USER_ID", "bar")
	assert.Nil(t, AppIDStrategy(fs))

	os.Setenv("VAULT_APP_ID", "foo")
	os.Setenv("VAULT_USER_ID", "bar")
	auth := AppIDStrategy(fs)
	assert.Equal(t, "foo", auth.body["app_id"])
	assert.Equal(t, "bar", auth.body["user_id"])
	assert.Equal(t, "app-id", auth.mount)

	os.Setenv("VAULT_APP_ID", "baz")
	os.Setenv("VAULT_USER_ID", "qux")
	os.Setenv("VAULT_AUTH_APP_ID_MOUNT", "quux")
	auth = AppIDStrategy(fs)
	assert.Equal(t, "baz", auth.body["app_id"])
	assert.Equal(t, "qux", auth.body["user_id"])
	assert.Equal(t, "quux", auth.mount)
}

func TestAppRoleStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_ROLE_ID")
	defer os.Unsetenv("VAULT_SECRET_ID")
	defer os.Unsetenv("VAULT_AUTH_APPROLE_MOUNT")
	defer os.Unsetenv("VAULT_ROLE_ID_FILE")
	defer os.Unsetenv("VAULT_SECRET_ID_FILE")
	fs := memfs.Create()
	os.Unsetenv("VAULT_ROLE_ID")
	os.Unsetenv("VAULT_SECRET_ID")
	assert.Nil(t, AppRoleStrategy(fs))

	os.Setenv("VAULT_ROLE_ID", "foo")
	assert.Nil(t, AppRoleStrategy(fs))

	os.Unsetenv("VAULT_ROLE_ID")
	os.Setenv("VAULT_SECRET_ID", "bar")
	assert.Nil(t, AppRoleStrategy(fs))

	os.Setenv("VAULT_ROLE_ID", "foo")
	os.Setenv("VAULT_SECRET_ID", "bar")
	auth := AppRoleStrategy(fs)
	assert.Equal(t, "foo", auth.body["role_id"])
	assert.Equal(t, "bar", auth.body["secret_id"])
	assert.Equal(t, "approle", auth.mount)

	os.Setenv("VAULT_ROLE_ID", "baz")
	os.Setenv("VAULT_SECRET_ID", "qux")
	os.Setenv("VAULT_AUTH_APPROLE_MOUNT", "quux")
	auth = AppRoleStrategy(fs)
	assert.Equal(t, "baz", auth.body["role_id"])
	assert.Equal(t, "qux", auth.body["secret_id"])
	assert.Equal(t, "quux", auth.mount)
}

func TestGitHubStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_AUTH_GITHUB_TOKEN")
	defer os.Unsetenv("VAULT_AUTH_GITHUB_MOUNT")
	defer os.Unsetenv("VAULT_AUTH_GITHUB_TOKEN_FILE")

	fs := memfs.Create()

	os.Unsetenv("VAULT_AUTH_GITHUB_TOKEN")
	assert.Nil(t, GitHubStrategy(fs))

	os.Setenv("VAULT_AUTH_GITHUB_TOKEN", "foo")
	auth := GitHubStrategy(fs)
	assert.Equal(t, "foo", auth.body["token"])
	assert.Equal(t, "github", auth.mount)

	os.Setenv("VAULT_AUTH_GITHUB_MOUNT", "bar")
	auth = GitHubStrategy(fs)
	assert.Equal(t, "foo", auth.body["token"])
	assert.Equal(t, "bar", auth.mount)
}

func TestUserPassStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_AUTH_USERNAME")
	defer os.Unsetenv("VAULT_AUTH_PASSWORD")
	defer os.Unsetenv("VAULT_AUTH_USERPASS_MOUNT")
	defer os.Unsetenv("VAULT_AUTH_USERNAME_FILE")
	defer os.Unsetenv("VAULT_AUTH_PASSWORD_FILE")

	fs := memfs.Create()

	assert.Nil(t, UserPassStrategy(fs))

	os.Setenv("VAULT_AUTH_USERNAME", "foo")
	assert.Nil(t, UserPassStrategy(fs))

	os.Unsetenv("VAULT_AUTH_USERNAME")
	os.Setenv("VAULT_AUTH_PASSWORD", "bar")
	assert.Nil(t, UserPassStrategy(fs))

	os.Setenv("VAULT_AUTH_USERNAME", "foo")
	os.Setenv("VAULT_AUTH_PASSWORD", "bar")
	auth := UserPassStrategy(fs)
	assert.Equal(t, "/v1/auth/userpass/login/foo", auth.path)
	assert.Equal(t, "bar", auth.body["password"])
	assert.Equal(t, "userpass", auth.mount)

	os.Setenv("VAULT_AUTH_USERNAME", "foo")
	os.Setenv("VAULT_AUTH_PASSWORD", "bar")
	os.Setenv("VAULT_AUTH_USERPASS_MOUNT", "baz")
	auth = UserPassStrategy(fs)
	assert.Equal(t, "/v1/auth/baz/login/foo", auth.path)
	assert.Equal(t, "bar", auth.body["password"])
	assert.Equal(t, "baz", auth.mount)
}
