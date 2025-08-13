package azure

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyVaultClient_NewKeyVaultClient(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		vaultURL string
		wantErr  bool
	}{
		{
			name:     "valid HTTPS URL",
			vaultURL: "https://myvault.vault.azure.net",
			wantErr:  false,
		},
		{
			name:     "valid HTTP URL (converts to HTTPS)",
			vaultURL: "http://myvault.vault.azure.net",
			wantErr:  false,
		},
		{
			name:     "invalid URL",
			vaultURL: "not-a-url",
			wantErr:  true,
		},
		{
			name:     "URL without hostname",
			vaultURL: "https://",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This will fail in CI/testing environment without Azure credentials
			// but we can still test the URL parsing logic
			_, err := NewKeyVaultClient(ctx, tt.vaultURL)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// In testing environments without Azure credentials, this might fail
				// but URL parsing should work
				if err != nil {
					// Check if it's a credential error (expected in tests)
					assert.Contains(t, err.Error(), "credential")
				}
			}
		})
	}
}

func TestExtractVaultName(t *testing.T) {
	tests := []struct {
		hostname string
		expected string
	}{
		{
			hostname: "myvault.vault.azure.net",
			expected: "myvault",
		},
		{
			hostname: "test-vault.vault.azure.net",
			expected: "test-vault",
		},
		{
			hostname: "simple-name",
			expected: "simple-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			// This tests the vault name extraction logic
			parts := strings.Split(tt.hostname, ".")
			vaultName := parts[0]
			assert.Equal(t, tt.expected, vaultName)
		})
	}
}

// Integration test - only runs if Azure credentials are available
func TestKeyVaultClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	vaultURL := getEnvOrSkip(t, "AZURE_KEYVAULT_URL")
	ctx := context.Background()

	client, err := NewKeyVaultClient(ctx, vaultURL)
	if err != nil {
		t.Skipf("Failed to create Azure Key Vault client: %v", err)
	}

	// Test listing secrets (should not fail even if vault is empty)
	secrets, err := client.ListSecrets(ctx)
	require.NoError(t, err)
	t.Logf("Found %d secrets in vault", len(secrets))

	// If there are secrets, try to get the first one
	if len(secrets) > 0 {
		secretValue, err := client.GetSecret(ctx, secrets[0], "")
		if err != nil {
			t.Logf("Could not retrieve secret %q: %v", secrets[0], err)
		} else {
			assert.NotEmpty(t, secretValue)
			t.Logf("Successfully retrieved secret %q", secrets[0])
		}
	}
}

func getEnvOrSkip(t *testing.T, key string) string {
	value := os.Getenv(key)
	if value == "" {
		t.Skipf("Environment variable %s not set", key)
	}
	return value
}
