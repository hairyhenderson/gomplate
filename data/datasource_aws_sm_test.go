package data

import (
	"context"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DummyAWSSecretsManagerSecretGetter - test double
type DummyAWSSecretsManagerSecretGetter struct {
	t                  *testing.T
	secretValut        *secretsmanager.GetSecretValueOutput
	err                awserr.Error
	mockGetSecretValue func(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
}

func (d DummyAWSSecretsManagerSecretGetter) GetSecretValueWithContext(_ context.Context, input *secretsmanager.GetSecretValueInput, _ ...request.Option) (*secretsmanager.GetSecretValueOutput, error) {
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

	tplain := map[string]interface{}{"type": "text/plain"}

	data := []struct {
		eParams map[string]interface{}
		u       string
		ePath   string
		args    string
	}{
		{u: "noddy", ePath: "noddy"},
		{u: "base", ePath: "base/extra", args: "extra"},
		{u: "/foo/", ePath: "/foo/extra", args: "/extra"},
		{u: "aws+sm:///foo", ePath: "/foo/bar", args: "bar"},
		{u: "aws+sm:foo", ePath: "foo"},
		{u: "aws+sm:foo/bar", ePath: "foo/bar"},
		{u: "aws+sm:/foo/bar", ePath: "/foo/bar"},
		{u: "aws+sm:foo", ePath: "foo/baz", args: "baz"},
		{u: "aws+sm:foo/bar", ePath: "foo/bar/baz", args: "baz"},
		{u: "aws+sm:/foo/bar", ePath: "/foo/bar/baz", args: "baz"},
		{u: "aws+sm:///foo", ePath: "/foo/dir/", args: "dir/"},
		{u: "aws+sm:///foo/", ePath: "/foo/"},
		{u: "aws+sm:///foo/", ePath: "/foo/baz", args: "baz"},
		{eParams: tplain, u: "aws+sm:foo?type=text/plain", ePath: "foo/baz", args: "baz"},
		{eParams: tplain, u: "aws+sm:foo/bar?type=text/plain", ePath: "foo/bar/baz", args: "baz"},
		{eParams: tplain, u: "aws+sm:/foo/bar?type=text/plain", ePath: "/foo/bar/baz", args: "baz"},
		{
			eParams: map[string]interface{}{
				"type":  "application/json",
				"param": "quux",
			},
			u:     "aws+sm:/foo/bar?type=text/plain",
			ePath: "/foo/bar/baz/qux",
			args:  "baz/qux?type=application/json&param=quux",
		},
	}

	for _, d := range data {
		args := []string{d.args}
		if d.args == "" {
			args = nil
		}
		params, p, err := parseDatasourceURLArgs(mustParseURL(d.u), args...)
		require.NoError(t, err)
		if d.eParams == nil {
			assert.Empty(t, params)
		} else {
			assert.EqualValues(t, d.eParams, params)
		}
		assert.Equal(t, d.ePath, p)
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

	_, err := readAWSSecretsManager(context.Background(), s, "/bar")
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

	_, err := readAWSSecretsManager(context.Background(), s, "/bar", "/foo", "/bla")
	assert.False(t, calledOk)
	assert.Error(t, err)
}

func TestAWSSecretsManager_GetParameterMissing(t *testing.T) {
	expectedErr := awserr.New("ParameterNotFound", "Test of error message", nil)
	s := simpleAWSSecretsManagerSourceHelper(DummyAWSSecretsManagerSecretGetter{
		t:   t,
		err: expectedErr,
	})

	_, err := readAWSSecretsManager(context.Background(), s, "")
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

	output, err := readAWSSecretsManager(context.Background(), s, "/bar")
	assert.True(t, calledOk)
	require.NoError(t, err)
	assert.Equal(t, []byte("blub"), output)
}

func TestAWSSecretsManager_ReadSecretBinary(t *testing.T) {
	calledOk := false
	s := simpleAWSSecretsManagerSourceHelper(DummyAWSSecretsManagerSecretGetter{
		t: t,
		mockGetSecretValue: func(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
			assert.Equal(t, "/foo/bar", *input.SecretId)
			calledOk = true
			return &secretsmanager.GetSecretValueOutput{SecretBinary: []byte("supersecret")}, nil
		},
	})

	output, err := readAWSSecretsManager(context.Background(), s, "/bar")
	assert.True(t, calledOk)
	require.NoError(t, err)
	assert.Equal(t, []byte("supersecret"), output)
}
