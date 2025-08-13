package azure

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

// KeyVaultClient is a wrapper around the Azure Key Vault client
type KeyVaultClient struct {
	client *azsecrets.Client
	vault  string
}

// NewKeyVaultClient creates a new Azure Key Vault client
func NewKeyVaultClient(ctx context.Context, vaultURL string) (*KeyVaultClient, error) {
	// Parse and validate the vault URL
	u, err := url.Parse(vaultURL)
	if err != nil {
		return nil, fmt.Errorf("invalid vault URL %q: %w", vaultURL, err)
	}

	if u.Host == "" {
		return nil, fmt.Errorf("vault URL must include hostname: %q", vaultURL)
	}

	// Ensure the URL uses HTTPS and has the correct format
	if u.Scheme != "https" {
		u.Scheme = "https"
	}

	// Extract vault name from hostname (e.g., "myvault.vault.azure.net" -> "myvault")
	vaultName := strings.Split(u.Host, ".")[0]

	// Use DefaultAzureCredential for authentication
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	// Create the Key Vault client
	client, err := azsecrets.NewClient(u.String(), cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Key Vault client: %w", err)
	}

	return &KeyVaultClient{
		client: client,
		vault:  vaultName,
	}, nil
}

// GetSecret retrieves a secret from Azure Key Vault
func (kv *KeyVaultClient) GetSecret(ctx context.Context, secretName string, version string) (string, error) {
	// Prepare the get secret options
	opts := &azsecrets.GetSecretOptions{}

	// If no version is specified, get the latest version
	var resp azsecrets.GetSecretResponse
	var err error

	if version != "" {
		resp, err = kv.client.GetSecret(ctx, secretName, version, opts)
	} else {
		resp, err = kv.client.GetSecret(ctx, secretName, "", opts)
	}

	if err != nil {
		return "", fmt.Errorf("failed to get secret %q: %w", secretName, err)
	}

	if resp.Value == nil {
		return "", fmt.Errorf("secret %q has no value", secretName)
	}

	return *resp.Value, nil
}

// ListSecrets lists all secrets in the Azure Key Vault
func (kv *KeyVaultClient) ListSecrets(ctx context.Context) ([]string, error) {
	var secrets []string

	pager := kv.client.NewListSecretPropertiesPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %w", err)
		}

		for _, secret := range page.Value {
			if secret.ID != nil {
				// Extract secret name from the full ID
				// ID format: https://vault.vault.azure.net/secrets/secretname
				idStr := string(*secret.ID)
				parts := strings.Split(idStr, "/")
				if len(parts) >= 5 {
					secretName := parts[len(parts)-1] // Last part is the secret name
					secrets = append(secrets, secretName)
				}
			}
		}
	}

	return secrets, nil
}
