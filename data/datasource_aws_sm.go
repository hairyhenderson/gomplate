package data

import (
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/pkg/errors"

	gaws "github.com/hairyhenderson/gomplate/aws"
)

// awsSecretsManagerGetter - A subset of Secrets Manager API for use in unit testing
type awsSecretsManagerGetter interface {
	GetSecretValue(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
}

func parseAWSSecretsManagerArgs(origPath string, args ...string) (paramPath string, err error) {
	paramPath = origPath
	if len(args) >= 1 {
		paramPath = path.Join(paramPath, args[0])
	}

	if len(args) >= 2 {
		err = errors.New("Maximum two arguments to aws+sm datasource: alias, extraPath")
	}
	return
}

func readAWSSecretsManager(source *Source, args ...string) (output []byte, err error) {
	if source.awsSecretsManager == nil {
		source.awsSecretsManager = secretsmanager.New(gaws.SDKSession())
	}

	paramPath, err := parseAWSSecretsManagerArgs(source.URL.Path, args...)
	if err != nil {
		return nil, err
	}

	source.mediaType = jsonMimetype
	return readAWSSecretsManagerParam(source, paramPath)
}

func readAWSSecretsManagerParam(source *Source, paramPath string) ([]byte, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(paramPath),
	}

	response, err := source.awsSecretsManager.GetSecretValue(input)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading aws+sm from AWS using GetSecretValue with input %v", input)
	}

	return toJSONBytes(response)
}
