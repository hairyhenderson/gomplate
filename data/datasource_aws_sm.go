package data

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	gaws "github.com/hairyhenderson/gomplate/v4/aws"
)

// awsSecretsManagerGetter - A subset of Secrets Manager API for use in unit testing
type awsSecretsManagerGetter interface {
	GetSecretValueWithContext(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error)
}

func parseDatasourceURLArgs(sourceURL *url.URL, args ...string) (params map[string]interface{}, p string, err error) {
	if len(args) >= 2 {
		err = fmt.Errorf("maximum two arguments to %s datasource: alias, extraPath (found %d)",
			sourceURL.Scheme, len(args))
		return nil, "", err
	}

	p = sourceURL.Path
	params = make(map[string]interface{})
	for key, val := range sourceURL.Query() {
		params[key] = strings.Join(val, " ")
	}

	if p == "" && sourceURL.Opaque != "" {
		p = sourceURL.Opaque
	}

	if len(args) == 1 {
		parsed, err := url.Parse(args[0])
		if err != nil {
			return nil, "", err
		}

		if parsed.Path != "" {
			p = path.Join(p, parsed.Path)
			if strings.HasSuffix(parsed.Path, "/") {
				p += "/"
			}
		}

		for key, val := range parsed.Query() {
			params[key] = strings.Join(val, " ")
		}
	}
	return params, p, nil
}

func readAWSSecretsManager(ctx context.Context, source *Source, args ...string) ([]byte, error) {
	if source.awsSecretsManager == nil {
		source.awsSecretsManager = secretsmanager.New(gaws.SDKSession())
	}

	_, paramPath, err := parseDatasourceURLArgs(source.URL, args...)
	if err != nil {
		return nil, err
	}

	return readAWSSecretsManagerParam(ctx, source, paramPath)
}

func readAWSSecretsManagerParam(ctx context.Context, source *Source, paramPath string) ([]byte, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(paramPath),
	}

	response, err := source.awsSecretsManager.GetSecretValueWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("reading aws+sm source %q: %w", source.Alias, err)
	}

	if response.SecretString != nil {
		return []byte(*response.SecretString), nil
	}

	return response.SecretBinary, nil
}
