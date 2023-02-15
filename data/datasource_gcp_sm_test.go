package data

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/googleapis/gax-go/v2/apierror"
	"github.com/stretchr/testify/assert"
)

// DummyGCPSecretsManagerSecretGetter - test double
type DummyGCPSecretsManagerSecretGetter struct {
	t                       *testing.T
	accessVersionResponse   *secretmanagerpb.AccessSecretVersionResponse
	secretVersion           *secretmanagerpb.SecretVersion
	err                     *apierror.APIError
	mockAccessSecretVersion func(req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error)
	mockGetSecretVersion    func(req *secretmanagerpb.GetSecretVersionRequest) (*secretmanagerpb.SecretVersion, error)
}

func (d DummyGCPSecretsManagerSecretGetter) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	if d.mockAccessSecretVersion != nil {
		output, err := d.mockAccessSecretVersion(req)
		return output, err
	}
	if d.err != nil {
		return nil, d.err
	}
	assert.NotNil(d.t, d.accessVersionResponse, "Must provide a param if no error!")
	return d.accessVersionResponse, nil
}
func (d DummyGCPSecretsManagerSecretGetter) GetSecretVersion(ctx context.Context, req *secretmanagerpb.GetSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
	if d.mockGetSecretVersion != nil {
		output, err := d.mockGetSecretVersion(req)
		return output, err
	}
	if d.err != nil {
		return nil, d.err
	}
	assert.NotNil(d.t, d.secretVersion, "Must provide a param if no error!")
	return d.secretVersion, nil
}

func simpleGCPSecretsManagerSourceHelper(dummyGetter gcpSecretsManagerGetter) *Source {
	return &Source{
		Alias: "foo",
		URL: &url.URL{
			Scheme: "gcp+sm",
			Path:   "/foo",
		},
		gcpSecretsManager: dummyGetter,
	}
}

func TestGCPSecretsManager_ParseGCPSecretsManagerArgs(t *testing.T) {
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
		{u: "gcp+sm:///foo", ePath: "/foo/bar", args: "bar"},
		{u: "gcp+sm:foo", ePath: "foo"},
		{u: "gcp+sm:foo/bar", ePath: "foo/bar"},
		{u: "gcp+sm:/foo/bar", ePath: "/foo/bar"},
		{u: "gcp+sm:foo", ePath: "foo/baz", args: "baz"},
		{u: "gcp+sm:foo/bar", ePath: "foo/bar/baz", args: "baz"},
		{u: "gcp+sm:/foo/bar", ePath: "/foo/bar/baz", args: "baz"},
		{u: "gcp+sm:///foo", ePath: "/foo/dir/", args: "dir/"},
		{u: "gcp+sm:///foo/", ePath: "/foo/"},
		{u: "gcp+sm:///foo/", ePath: "/foo/baz", args: "baz"},
		{eParams: tplain, u: "gcp+sm:foo?type=text/plain", ePath: "foo/baz", args: "baz"},
		{eParams: tplain, u: "gcp+sm:foo/bar?type=text/plain", ePath: "foo/bar/baz", args: "baz"},
		{eParams: tplain, u: "gcp+sm:/foo/bar?type=text/plain", ePath: "/foo/bar/baz", args: "baz"},
		{
			eParams: map[string]interface{}{
				"type":  "application/json",
				"param": "quux",
			},
			u:     "gcp+sm:/foo/bar?type=text/plain",
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
		assert.NoError(t, err)
		if d.eParams == nil {
			assert.Empty(t, params)
		} else {
			assert.EqualValues(t, d.eParams, params)
		}
		assert.Equal(t, d.ePath, p)
	}
}

func TestGCPSecretsManager_GetParameterSetup(t *testing.T) {
	calledOk := false
	s := simpleGCPSecretsManagerSourceHelper(DummyGCPSecretsManagerSecretGetter{
		t: t,
		mockAccessSecretVersion: func(req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
			assert.Equal(t, "foo/bar", req.Name)
			calledOk = true
			return &secretmanagerpb.AccessSecretVersionResponse{
				Name: req.Name,
				Payload: &secretmanagerpb.SecretPayload{
					Data: []byte("blub"),
				},
			}, nil

		},
		mockGetSecretVersion: func(req *secretmanagerpb.GetSecretVersionRequest) (*secretmanagerpb.SecretVersion, error) {
			assert.Equal(t, "foo/bar", req.Name)
			calledOk = true
			return &secretmanagerpb.SecretVersion{
				Name: req.Name,
			}, nil
		},
	})

	_, err := readGCPSecretsManager(context.Background(), s, "/bar")
	assert.True(t, calledOk)
	assert.Nil(t, err)
}

func TestGCPSecretsManager_GetParameterSetupWrongArgs(t *testing.T) {
	calledOk := false
	s := simpleGCPSecretsManagerSourceHelper(DummyGCPSecretsManagerSecretGetter{
		t: t,
		mockAccessSecretVersion: func(req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
			assert.Equal(t, "/foo/bar", req.Name)
			calledOk = true
			return &secretmanagerpb.AccessSecretVersionResponse{
				Name: req.Name,
				Payload: &secretmanagerpb.SecretPayload{
					Data: []byte("blub"),
				},
			}, nil

		},
		mockGetSecretVersion: func(req *secretmanagerpb.GetSecretVersionRequest) (*secretmanagerpb.SecretVersion, error) {
			assert.Equal(t, "/foo/bar", req.Name)
			calledOk = true
			return &secretmanagerpb.SecretVersion{
				Name: req.Name,
			}, nil
		},
	})

	_, err := readGCPSecretsManager(context.Background(), s, "/bar", "/foo", "/bla")
	assert.False(t, calledOk)
	assert.Error(t, err)
}

func TestGCPSecretsManager_GetParameterMissing(t *testing.T) {
	stat, _ := status.New(codes.InvalidArgument, "Invalid resource field value in the request.").WithDetails()
	expectedErr, _ := apierror.FromError(stat.Err())
	s := simpleGCPSecretsManagerSourceHelper(DummyGCPSecretsManagerSecretGetter{
		t:   t,
		err: expectedErr,
	})

	_, err := readGCPSecretsManager(context.Background(), s, "")
	assert.Error(t, err, "Test of error message")
}
