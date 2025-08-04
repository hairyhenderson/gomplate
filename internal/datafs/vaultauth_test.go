package datafs

import (
    "io/fs"
    "os"
    "testing"

    authk8s "github.com/hashicorp/vault/api/auth/kubernetes"
)

type dummyFS struct{}

// implement ReadFile if needed
func (d dummyFS) Open(name string) (fs.File, error) { return nil, fs.ErrNotExist }

func TestEnvKubernetesAuthAdapter_NoRole(t *testing.T) {
    fsys := fs.OS // fallback, role unset
    os.Unsetenv("VAULT_AUTH_K8S_ROLE")
    method := envKubernetesAuthAdapter(fsys)
    if method != nil {
        t.Fatal("Expected nil adapter when VAULT_AUTH_K8S_ROLE is unset")
    }
}

func TestEnvKubernetesAuthAdapter_WithRole(t *testing.T) {
    os.Setenv("VAULT_AUTH_K8S_ROLE", "test-role")
    os.Setenv("VAULT_AUTH_K8S_MOUNT", "myk8s")
    tempFile := "/tmp/test-jwt.token"
    os.WriteFile(tempFile, []byte("dummy-jwt"), 0o600)
    os.Setenv("VAULT_AUTH_K8S_JWT_PATH", tempFile)

    method := envKubernetesAuthAdapter(fs.OS)
    if method == nil {
        t.Fatal("Expected non-nil adapter when VAULT_AUTH_K8S_ROLE is set")
    }

    _, ok := method.(*authk8s.KubernetesAuth)
    if !ok {
        t.Fatalf("Expected KubernetesAuth type got %T", method)
    }
}
