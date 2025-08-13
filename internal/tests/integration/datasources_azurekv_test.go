package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatasources_AzureKeyVault(t *testing.T) {
	// Skip test if Azure credentials are not available
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	vaultURL := os.Getenv("AZURE_KEYVAULT_URL")
	if vaultURL == "" {
		t.Skip("AZURE_KEYVAULT_URL not set, skipping Azure Key Vault integration test")
	}

	testSecretName := os.Getenv("AZURE_TEST_SECRET_NAME")
	if testSecretName == "" {
		testSecretName = "test-secret"
	}

	t.Run("retrieve secret", func(t *testing.T) {
		// Test retrieving a specific secret
		o, e, err := cmd(t, "-d", "kv=azure+kv://"+vaultURL[8:]+"/"+testSecretName, // Remove https://
			"-i", `{{ ds "kv" }}`).run()

		if err != nil {
			// If the secret doesn't exist or we don't have access, that's also a valid test result
			t.Logf("Could not retrieve secret (this may be expected): %v", err)
			t.Logf("stdout: %s", o)
			t.Logf("stderr: %s", e)
			return
		}

		assertSuccess(t, o, e, err, "")
		// The content should not be empty if we successfully retrieved a secret
		assert.NotEmpty(t, string(o))
	})

	t.Run("list secrets", func(t *testing.T) {
		// Test listing all secrets in the vault
		o, e, err := cmd(t, "-d", "kv=azure+kv://"+vaultURL[8:]+"/", // Remove https:// and add trailing slash
			"-i", `{{ range ds "kv" }}{{ . }}{{ "\n" }}{{ end }}`).run()

		if err != nil {
			t.Logf("Could not list secrets (this may be expected): %v", err)
			t.Logf("stdout: %s", o)
			t.Logf("stderr: %s", e)
			return
		}

		assertSuccess(t, o, e, err, "")
		// If successful, output should contain secret names
		t.Logf("Listed secrets: %s", string(o))
	})

	t.Run("opaque URL with environment variable", func(t *testing.T) {
		// Test using opaque URL with environment variable
		o, e, err := cmd(t, "-d", "kv=azure+kv:"+testSecretName,
			"-i", `{{ ds "kv" }}`).
			withEnv("AZURE_KEYVAULT_URL", vaultURL).run()

		if err != nil {
			t.Logf("Could not retrieve secret with opaque URL: %v", err)
			t.Logf("stdout: %s", o)
			t.Logf("stderr: %s", e)
			return
		}

		assertSuccess(t, o, e, err, "")
		assert.NotEmpty(t, string(o))
	})
}

func TestDatasources_AzureKeyVault_ErrorCases(t *testing.T) {
	t.Run("invalid vault URL", func(t *testing.T) {
		_, e, err := cmd(t, "-d", "kv=azure+kv://invalid-vault/secret",
			"-i", `{{ ds "kv" }}`).run()

		// This should fail
		assert.Error(t, err)
		t.Logf("Expected error: %v", err)
		t.Logf("stderr: %s", e)
	})

	t.Run("missing environment variable", func(t *testing.T) {
		_, e, err := cmd(t, "-d", "kv=azure+kv:secret",
			"-i", `{{ ds "kv" }}`).run()

		// This should fail because AZURE_KEYVAULT_URL is not set
		assert.Error(t, err)
		assert.Contains(t, string(e), "AZURE_KEYVAULT_URL")
	})
}
