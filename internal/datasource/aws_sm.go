package datasource

import (
	"context"
	"net/url"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/pkg/errors"

	gaws "github.com/hairyhenderson/gomplate/v3/aws"
)

// awsSecretsManagerGetter - A subset of Secrets Manager API for use in unit testing
type awsSecretsManagerGetter interface {
	GetSecretValueWithContext(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error)
}

// AWSSecretsManager -
type AWSSecretsManager struct {
	awsSecretsManager awsSecretsManagerGetter
}

var _ Reader = (*AWSSecretsManager)(nil)

func (a *AWSSecretsManager) Read(ctx context.Context, url *url.URL, args ...string) (data *Data, err error) {
	if a.awsSecretsManager == nil {
		a.awsSecretsManager = secretsmanager.New(gaws.SDKSession())
	}

	data = newData(url, args)

	paramPath, err := a.parseArgs(url.Path, args...)
	if err != nil {
		return nil, err
	}

	data.Bytes, err = a.getSecret(ctx, paramPath)
	return data, err
}

func (a *AWSSecretsManager) getSecret(ctx context.Context, paramPath string) ([]byte, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(paramPath),
	}

	response, err := a.awsSecretsManager.GetSecretValueWithContext(ctx, input)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading aws+sm from AWS using GetSecretValue with input %v", input)
	}

	return []byte(aws.StringValue(response.SecretString)), nil
}

func (a *AWSSecretsManager) parseArgs(origPath string, args ...string) (paramPath string, err error) {
	paramPath = origPath
	if len(args) >= 1 {
		paramPath = path.Join(paramPath, args[0])
	}

	if len(args) >= 2 {
		err = errors.New("Maximum two arguments to aws+sm datasource: alias, extraPath")
	}
	return
}
