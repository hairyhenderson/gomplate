package datafs

import (
	"os"
	"testing"
	"testing/fstest"

	authk8s "github.com/hashicorp/vault/api/auth/kubernetes"
)

func TestEnvKubernetesAuthAdapter_NoRole(t *testing.T) {
	t.Setenv("VAULT_AUTH_K8S_ROLE", "") // Make env var recoverable after test
	os.Unsetenv("VAULT_AUTH_K8S_ROLE")  // Force `os.Unsetenv` as there is no `t.Unsetenv`
	method := envKubernetesAuthAdapter(fstest.MapFS{})
	if method != nil {
		t.Fatal("Expected nil adapter when VAULT_AUTH_K8S_ROLE is unset")
	}
}

func TestEnvKubernetesAuthAdapter_WithRole(t *testing.T) {
	t.Setenv("VAULT_AUTH_K8S_ROLE", "test-role")
	t.Setenv("VAULT_AUTH_K8S_MOUNT", "myk8s")
	tempFile := "/tmp/test-jwt.token"
	t.Setenv("VAULT_AUTH_K8S_JWT_PATH", tempFile)

	fsys := &fstest.MapFS{
		tempFile: {
			Data: []byte("dummy-jwt"),
			Mode: 0o600,
		},
	}
	method := envKubernetesAuthAdapter(fsys)
	if method == nil {
		t.Fatal("Expected non-nil adapter when VAULT_AUTH_K8S_ROLE is set")
	}

	_, ok := method.(*authk8s.KubernetesAuth)
	if !ok {
		t.Fatalf("Expected KubernetesAuth type got %T", method)
	}
}
