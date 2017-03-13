package vault

import (
	"fmt"
	"os"
)

// AppIDAuthStrategy - an AuthStrategy that uses Vault's app-id authentication backend.
type AppIDAuthStrategy struct {
	*Strategy
	AppID  string `json:"app_id"`
	UserID string `json:"user_id"`
}

// NewAppIDAuthStrategy - create an AuthStrategy that uses Vault's app-id auth
// backend.
func NewAppIDAuthStrategy() *AppIDAuthStrategy {
	appID := os.Getenv("VAULT_APP_ID")
	userID := os.Getenv("VAULT_USER_ID")
	mount := os.Getenv("VAULT_AUTH_APP_ID_MOUNT")
	if mount == "" {
		mount = "app-id"
	}
	if appID != "" && userID != "" {
		return &AppIDAuthStrategy{&Strategy{mount, nil}, appID, userID}
	}
	return nil
}

func (a *AppIDAuthStrategy) String() string {
	return fmt.Sprintf("app-id: %s, user-id: %s, mount: %s", a.AppID, a.UserID, a.Mount)
}
