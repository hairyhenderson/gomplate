package datasources

import (
	"context"
	"io/ioutil"
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
	secretValue        *secretsmanager.GetSecretValueOutput
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
	assert.NotNil(d.t, d.secretValue, "Must provide a param if no error!")
	return d.secretValue, nil
}

func TestAWSSecretsManager_GetParameterMissing(t *testing.T) {
	expectedErr := awserr.New("ParameterNotFound", "Test of error message", nil)
	ctx := context.Background()
	r := &awsSecretsManagerRequester{DummyAWSSecretsManagerSecretGetter{
		t:   t,
		err: expectedErr,
	}}

	_, err := r.Request(ctx, mustParseURL("aws+sm:///foo"), nil)
	assert.Error(t, err, "Test of error message")
}

func TestAWSSecretsManager_ReadSecret(t *testing.T) {
	ctx := context.Background()
	r := &awsSecretsManagerRequester{DummyAWSSecretsManagerSecretGetter{
		t: t,
		mockGetSecretValue: func(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
			assert.Equal(t, "/foo/bar", *input.SecretId)
			return &secretsmanager.GetSecretValueOutput{SecretString: aws.String("blub")}, nil
		},
	}}

	resp, err := r.Request(ctx, mustParseURL("aws+sm:///foo/bar"), nil)
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, []byte("blub"), b)
}
