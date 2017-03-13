package vault

import (
	"fmt"
	"os"
)

// GitHubAuthStrategy - an AuthStrategy that uses Vault's app-id authentication backend.
type GitHubAuthStrategy struct {
	*Strategy
	Token string `json:"token"`
}

// NewGitHubAuthStrategy - create an AuthStrategy that uses Vault's app-id auth
// backend.
func NewGitHubAuthStrategy() *GitHubAuthStrategy {
	mount := os.Getenv("VAULT_AUTH_GITHUB_MOUNT")
	if mount == "" {
		mount = "github"
	}
	token := os.Getenv("VAULT_AUTH_GITHUB_TOKEN")
	if token != "" {
		return &GitHubAuthStrategy{&Strategy{mount, nil}, token}
	}
	return nil
}

func (a *GitHubAuthStrategy) String() string {
	return fmt.Sprintf("token: %s, mount: %s", a.Token, a.Mount)
}
