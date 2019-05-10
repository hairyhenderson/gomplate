package data

import (
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

func (d DummyAWSSecretsManagerSecretGetter) GetSecretValue(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
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

func simpleAWSSecretsManagerSourceHelper(dummyGetter awsSecretsManagerGetter) *Source {
	return &Source{
		Alias: "foo",
		URL: &url.URL{
			Scheme: "aws+sm",
			Path:   "/foo",
		},
		awsSecretsManager: dummyGetter,
	}
}

func TestAWSSecretsManager_ParseArgsSimple(t *testing.T) {
	paramPath, err := parseAWSSecretsManagerArgs("noddy")
	assert.Equal(t, "noddy", paramPath)
	assert.Nil(t, err)
}

func TestAWSSecretsManager_ParseArgsAppend(t *testing.T) {
	paramPath, err := parseAWSSecretsManagerArgs("base", "extra")
	assert.Equal(t, "base/extra", paramPath)
	assert.Nil(t, err)
}

func TestAWSSecretsManager_ParseArgsAppend2(t *testing.T) {
	paramPath, err := parseAWSSecretsManagerArgs("/foo/", "/extra")
	assert.Equal(t, "/foo/extra", paramPath)
	assert.Nil(t, err)
}

func TestAWSSecretsManager_ParseArgsTooMany(t *testing.T) {
	_, err := parseAWSSecretsManagerArgs("base", "extra", "too many!")
	assert.Error(t, err)
}

func TestAWSSecretsManager_GetParameterSetup(t *testing.T) {
	calledOk := false
	s := simpleAWSSecretsManagerSourceHelper(DummyAWSSecretsManagerSecretGetter{
		t: t,
		mockGetSecretValue: func(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
			assert.Equal(t, "/foo/bar", *input.SecretId)
			calledOk = true
			return &secretsmanager.GetSecretValueOutput{SecretString: aws.String("blub")}, nil
		},
	})

	_, err := readAWSSecretsManager(s, "/bar")
	assert.True(t, calledOk)
	assert.Nil(t, err)
}

func TestAWSSecretsManager_GetParameterSetupWrongArgs(t *testing.T) {
	calledOk := false
	s := simpleAWSSecretsManagerSourceHelper(DummyAWSSecretsManagerSecretGetter{
		t: t,
		mockGetSecretValue: func(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
			assert.Equal(t, "/foo/bar", *input.SecretId)
			calledOk = true
			return &secretsmanager.GetSecretValueOutput{SecretString: aws.String("blub")}, nil
		},
	})

	_, err := readAWSSecretsManager(s, "/bar", "/foo", "/bla")
	assert.False(t, calledOk)
	assert.Error(t, err)
}

func TestAWSSecretsManager_GetParameterMissing(t *testing.T) {
	expectedErr := awserr.New("ParameterNotFound", "Test of error message", nil)
	s := simpleAWSSecretsManagerSourceHelper(DummyAWSSecretsManagerSecretGetter{
		t:   t,
		err: expectedErr,
	})

	_, err := readAWSSecretsManager(s, "")
	assert.Error(t, err, "Test of error message")
}

func TestAWSSecretsManager_ReadSecret(t *testing.T) {
	calledOk := false
	s := simpleAWSSecretsManagerSourceHelper(DummyAWSSecretsManagerSecretGetter{
		t: t,
		mockGetSecretValue: func(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
			assert.Equal(t, "/foo/bar", *input.SecretId)
			calledOk = true
			return &secretsmanager.GetSecretValueOutput{SecretString: aws.String("blub")}, nil
		},
	})

	output, err := readAWSSecretsManagerParam(s, "/foo/bar")
	assert.True(t, calledOk)
	assert.NoError(t, err)
	assert.Equal(t, []byte("blub"), output)
}
