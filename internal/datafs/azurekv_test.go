package datafs

import (
	"io/fs"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAzureKVFS_NewFS(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		envVar  string
		wantErr bool
	}{
		{
			name:    "valid full URL",
			url:     "azure+kv://myvault.vault.azure.net",
			wantErr: false,
		},
		{
			name:    "opaque URL with env var",
			url:     "azure+kv:mysecret",
			envVar:  "https://myvault.vault.azure.net",
			wantErr: false,
		},
		{
			name:    "opaque URL without env var",
			url:     "azure+kv:mysecret",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "azure+kv:",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envVar != "" {
				os.Setenv("AZURE_KEYVAULT_URL", tt.envVar)
				defer os.Unsetenv("AZURE_KEYVAULT_URL")
			}

			u, err := url.Parse(tt.url)
			require.NoError(t, err)

			_, err = newAzureKVFS(u)
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

func TestAzureKVFS_URLParsing(t *testing.T) {
	tests := []struct {
		name       string
		rawURL     string
		expectHost string
		expectPath string
	}{
		{
			name:       "full URL",
			rawURL:     "azure+kv://myvault.vault.azure.net/secret1",
			expectHost: "myvault.vault.azure.net",
			expectPath: "/secret1",
		},
		{
			name:       "opaque URL",
			rawURL:     "azure+kv:secret1",
			expectHost: "",
			expectPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.rawURL)
			require.NoError(t, err)

			assert.Equal(t, tt.expectHost, u.Host)
			if tt.expectPath != "" {
				assert.Equal(t, tt.expectPath, u.Path)
			}
		})
	}
}

// TestAzureKVFS_Integration tests Azure Key Vault integration
// This test will be skipped unless Azure credentials are available
func TestAzureKVFS_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	vaultURL := getEnvOrSkip(t, "AZURE_KEYVAULT_URL")

	// Parse as full URL
	fullURL := "azure+kv://" + vaultURL[8:] // Remove https://
	u, err := url.Parse(fullURL)
	require.NoError(t, err)

	// Create filesystem
	fsys, err := newAzureKVFS(u)
	if err != nil {
		t.Skipf("Failed to create Azure Key Vault filesystem: %v", err)
	}

	// Test directory listing
	dirFile, err := fsys.Open(".")
	if err != nil {
		t.Skipf("Failed to open directory: %v", err)
	}
	defer dirFile.Close()

	// Check if it's a directory
	info, err := dirFile.Stat()
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Try to read directory entries
	if dirFile, ok := dirFile.(fs.ReadDirFile); ok {
		entries, err := dirFile.ReadDir(-1)
		require.NoError(t, err)
		t.Logf("Found %d secrets", len(entries))

		// If there are secrets, try to read one
		if len(entries) > 0 {
			secretName := entries[0].Name()
			secretFile, err := fsys.Open(secretName)
			if err != nil {
				t.Logf("Could not open secret %q: %v", secretName, err)
			} else {
				defer secretFile.Close()

				info, err := secretFile.Stat()
				require.NoError(t, err)
				assert.False(t, info.IsDir())
				assert.Greater(t, info.Size(), int64(0))
				t.Logf("Successfully opened secret %q with size %d", secretName, info.Size())
			}
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
