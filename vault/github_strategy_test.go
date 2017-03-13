package vault

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGitHubAuthStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_AUTH_GITHUB_TOKEN")
	defer os.Unsetenv("VAULT_AUTH_GITHUB_MOUNT")

	os.Unsetenv("VAULT_AUTH_GITHUB_TOKEN")
	assert.Nil(t, NewGitHubAuthStrategy())

	os.Setenv("VAULT_AUTH_GITHUB_TOKEN", "foo")
	auth := NewGitHubAuthStrategy()
	assert.Equal(t, "foo", auth.Token)
	assert.Equal(t, "github", auth.Mount)

	os.Setenv("VAULT_AUTH_GITHUB_MOUNT", "bar")
	auth = NewGitHubAuthStrategy()
	assert.Equal(t, "foo", auth.Token)
	assert.Equal(t, "bar", auth.Mount)
}
