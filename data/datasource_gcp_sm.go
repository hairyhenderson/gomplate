package data

import (
	"context"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
)

// gcpSecretsManagerGetter - A subset of Secrets Manager API for use in unit testing
type gcpSecretsManagerGetter interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	GetSecretVersion(ctx context.Context, req *secretmanagerpb.GetSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
}

func readGCPSecretsManager(ctx context.Context, source *Source, args ...string) ([]byte, error) {
	if source.gcpSecretsManager == nil {
		client, err := secretmanager.NewClient(ctx)
		if err != nil {
			return nil, err
		}
		source.gcpSecretsManager = client
	}

	_, paramPath, err := parseDatasourceURLArgs(source.URL, args...)
	if err != nil {
		return nil, err
	}
	paramPath = strings.TrimLeft(paramPath, "/")

	vreq := secretmanagerpb.GetSecretVersionRequest{
		Name: paramPath,
	}
	version, err := source.gcpSecretsManager.GetSecretVersion(ctx, &vreq)
	if err != nil {
		return nil, err
	}

	req := secretmanagerpb.AccessSecretVersionRequest{
		Name: version.Name,
	}

	versionData, err := source.gcpSecretsManager.AccessSecretVersion(ctx, &req)
	if err != nil {
		return nil, err
	}

	return versionData.Payload.Data, nil
}
