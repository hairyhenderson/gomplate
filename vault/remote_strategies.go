package vault

import "os"

func createStrategy(mount string, body map[string]string, path ...string) *Strategy {
	for _, v := range body {
		if v == "" {
			return nil
		}
	}
	if len(path) > 0 {
		return &Strategy{mount: mount, body: body, path: path[0]}
	}
	return &Strategy{mount: mount, body: body}
}

// Would use env.go, but can't since it's in the main package...
func getenv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

// AppIDStrategy - app-id auth backend
func AppIDStrategy() *Strategy {
	mount := getenv("VAULT_AUTH_APP_ID_MOUNT", "app-id")
	return createStrategy(mount, map[string]string{
		"app_id":  os.Getenv("VAULT_APP_ID"),
		"user_id": os.Getenv("VAULT_USER_ID"),
	})
}

// AppRoleStrategy - approle auth backend
func AppRoleStrategy() *Strategy {
	mount := getenv("VAULT_AUTH_APPROLE_MOUNT", "approle")
	return createStrategy(mount, map[string]string{
		"role_id":   os.Getenv("VAULT_ROLE_ID"),
		"secret_id": os.Getenv("VAULT_SECRET_ID"),
	})
}

// GitHubStrategy - github auth backend
func GitHubStrategy() *Strategy {
	mount := getenv("VAULT_AUTH_GITHUB_MOUNT", "github")
	return createStrategy(mount, map[string]string{
		"token": os.Getenv("VAULT_AUTH_GITHUB_TOKEN"),
	})
}

// UserPassStrategy - userpass auth backend
func UserPassStrategy() *Strategy {
	username := os.Getenv("VAULT_AUTH_USERNAME")
	if username != "" {
		mount := getenv("VAULT_AUTH_USERPASS_MOUNT", "userpass")
		return createStrategy(mount, map[string]string{
			"password": os.Getenv("VAULT_AUTH_PASSWORD"),
		}, "/v1/auth/"+mount+"/login/"+username)
	}
	return nil
}
