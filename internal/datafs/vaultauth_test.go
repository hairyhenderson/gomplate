package datafs

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	authk8s "github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/stretchr/testify/require"
)

func TestEnvKubernetesAuthAdapter_NoRole(t *testing.T) {
	t.Setenv("VAULT_AUTH_K8S_ROLE", "") // Make env var recoverable after test
	os.Unsetenv("VAULT_AUTH_K8S_ROLE")  // Force `os.Unsetenv` as there is no `t.Unsetenv`

	require.Nil(t, envKubernetesAuthAdapter(fstest.MapFS{}), "Expected nil adapter when VAULT_AUTH_K8S_ROLE is unset")
}

func TestEnvKubernetesAuthAdapter_WithRole(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "test-jwt.token")

	realFs := os.DirFS("/")
	t.Setenv("VAULT_AUTH_K8S_ROLE", "test-role")
	t.Setenv("VAULT_AUTH_K8S_MOUNT", "myk8s")
	t.Setenv("VAULT_AUTH_K8S_JWT_PATH", tempFile)

	require.NoError(t, os.WriteFile(tempFile, []byte("dummy-jwt"), 0o600))

	require.IsType(t, &authk8s.KubernetesAuth{}, envKubernetesAuthAdapter(realFs))
}
