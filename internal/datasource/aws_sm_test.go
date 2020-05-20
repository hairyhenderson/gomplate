package datasource

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/stretchr/testify/assert"
)

// DummyAWSSecretsManagerSecretGetter - test double
type DummyAWSSecretsManagerSecretGetter struct {
	t                  *testing.T
	secretValut        *secretsmanager.GetSecretValueOutput
	err                awserr.Error
	mockGetSecretValue func(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
}

func (d DummyAWSSecretsManagerSecretGetter) GetSecretValueWithContext(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error) {
	if d.mockGetSecretValue != nil {
		output, err := d.mockGetSecretValue(input)
		return output, err
	}
	if d.err != nil {
		return nil, d.err
	}
	assert.NotNil(d.t, d.secretValut, "Must provide a param if no error!")
	return d.secretValut, nil
}

func TestAWSSecretsManager_ParseArgsSimple(t *testing.T) {
	a := &AWSSecretsManager{}
	paramPath, err := a.parseArgs("noddy")
	assert.Equal(t, "noddy", paramPath)
	assert.Nil(t, err)
}

func TestAWSSecretsManager_ParseArgsAppend(t *testing.T) {
	a := &AWSSecretsManager{}
	paramPath, err := a.parseArgs("base", "extra")
	assert.Equal(t, "base/extra", paramPath)
	assert.Nil(t, err)
}

func TestAWSSecretsManager_ParseArgsAppend2(t *testing.T) {
	a := &AWSSecretsManager{}
	paramPath, err := a.parseArgs("/foo/", "/extra")
	assert.Equal(t, "/foo/extra", paramPath)
	assert.Nil(t, err)
}

func TestAWSSecretsManager_ParseArgsTooMany(t *testing.T) {
	a := &AWSSecretsManager{}
	_, err := a.parseArgs("base", "extra", "too many!")
	assert.Error(t, err)
}

func TestAWSSecretsManager_GetParameterSetup(t *testing.T) {
	calledOk := false
	d := DummyAWSSecretsManagerSecretGetter{
		t: t,
		mockGetSecretValue: func(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
			assert.Equal(t, "/foo/bar", *input.SecretId)
			calledOk = true
			return &secretsmanager.GetSecretValueOutput{SecretString: aws.String("blub")}, nil
		},
	}

	ctx := context.Background()
	a := &AWSSecretsManager{awsSecretsManager: d}

	_, err := a.Read(ctx, mustParseURL("aws+sm:///foo"), "/bar")
	assert.True(t, calledOk)
	assert.Nil(t, err)
}

func TestAWSSecretsManager_GetParameterSetupWrongArgs(t *testing.T) {
	calledOk := false
	d := DummyAWSSecretsManagerSecretGetter{
		t: t,
		mockGetSecretValue: func(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
			assert.Equal(t, "/foo/bar", *input.SecretId)
			calledOk = true
			return &secretsmanager.GetSecretValueOutput{SecretString: aws.String("blub")}, nil
		},
	}

	ctx := context.Background()
	a := &AWSSecretsManager{awsSecretsManager: d}

	_, err := a.Read(ctx, mustParseURL("aws+sm:///foo"), "/bar", "/foo", "bla")
	assert.False(t, calledOk)
	assert.Error(t, err)
}

func TestAWSSecretsManager_GetParameterMissing(t *testing.T) {
	expectedErr := awserr.New("ParameterNotFound", "Test of error message", nil)
	d := DummyAWSSecretsManagerSecretGetter{
		t:   t,
		err: expectedErr,
	}

	ctx := context.Background()
	a := &AWSSecretsManager{awsSecretsManager: d}

	_, err := a.Read(ctx, mustParseURL("aws+sm:///foo"), "")
	assert.Error(t, err, "Test of error message")
}

func TestAWSSecretsManager_ReadSecret(t *testing.T) {
	calledOk := false
	d := DummyAWSSecretsManagerSecretGetter{
		t: t,
		mockGetSecretValue: func(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
			assert.Equal(t, "/foo/bar", *input.SecretId)
			calledOk = true
			return &secretsmanager.GetSecretValueOutput{SecretString: aws.String("blub")}, nil
		},
	}

	ctx := context.Background()
	a := &AWSSecretsManager{awsSecretsManager: d}

	output, err := a.getSecret(ctx, "/foo/bar")
	assert.True(t, calledOk)
	assert.NoError(t, err)
	assert.Equal(t, []byte("blub"), output)
}
