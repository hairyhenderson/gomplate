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

func TestAWSSecretsManager_ParseAWSSecretsManagerArgs(t *testing.T) {
	_, _, err := parseDatasourceURLArgs(mustParseURL("base"), "extra", "too many!")
	assert.Error(t, err)

	data := []struct {
		u              *url.URL
		args           []string
		expectedParams map[string]interface{}
		expectedPath   string
	}{
		{mustParseURL("noddy"), nil, nil, "noddy"},
		{mustParseURL("base"), []string{"extra"}, nil, "base/extra"},
		{mustParseURL("/foo/"), []string{"/extra"}, nil, "/foo/extra"},
		{mustParseURL("aws+sm:///foo"), []string{"bar"}, nil, "/foo/bar"},
		{mustParseURL("aws+sm:foo"), nil, nil, "foo"},
		{mustParseURL("aws+sm:foo/bar"), nil, nil, "foo/bar"},
		{mustParseURL("aws+sm:/foo/bar"), nil, nil, "/foo/bar"},
		{mustParseURL("aws+sm:foo"), []string{"baz"}, nil, "foo/baz"},
		{mustParseURL("aws+sm:foo/bar"), []string{"baz"}, nil, "foo/bar/baz"},
		{mustParseURL("aws+sm:/foo/bar"), []string{"baz"}, nil, "/foo/bar/baz"},
		{mustParseURL("aws+sm:///foo"), []string{"dir/"}, nil, "/foo/dir/"},
		{mustParseURL("aws+sm:///foo/"), nil, nil, "/foo/"},
		{mustParseURL("aws+sm:///foo/"), []string{"baz"}, nil, "/foo/baz"},

		{mustParseURL("aws+sm:foo?type=text/plain"), []string{"baz"},
			map[string]interface{}{"type": "text/plain"}, "foo/baz"},
		{mustParseURL("aws+sm:foo/bar?type=text/plain"), []string{"baz"},
			map[string]interface{}{"type": "text/plain"}, "foo/bar/baz"},
		{mustParseURL("aws+sm:/foo/bar?type=text/plain"), []string{"baz"},
			map[string]interface{}{"type": "text/plain"}, "/foo/bar/baz"},
		{
			mustParseURL("aws+sm:/foo/bar?type=text/plain"),
			[]string{"baz/qux?type=application/json&param=quux"},
			map[string]interface{}{
				"type":  "application/json",
				"param": "quux",
			},
			"/foo/bar/baz/qux",
		},
	}

	for _, d := range data {
		params, p, err := parseDatasourceURLArgs(d.u, d.args...)
		assert.NoError(t, err)
		if d.expectedParams == nil {
			assert.Empty(t, params)
		} else {
			assert.EqualValues(t, d.expectedParams, params)
		}
		assert.Equal(t, d.expectedPath, p)
	}
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
