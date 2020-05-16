package data

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/stretchr/testify/assert"
)

// DummyParamGetter - test double
type DummyParamGetter struct {
	t                *testing.T
	param            *ssm.Parameter
	params           []*ssm.Parameter
	err              awserr.Error
	mockGetParameter func(*ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
}

func (d DummyParamGetter) GetParameterWithContext(ctx context.Context, input *ssm.GetParameterInput, opts ...request.Option) (*ssm.GetParameterOutput, error) {
	if d.mockGetParameter != nil {
		output, err := d.mockGetParameter(input)
		return output, err
	}
	if d.err != nil {
		return nil, d.err
	}
	assert.NotNil(d.t, d.param, "Must provide a param if no error!")
	return &ssm.GetParameterOutput{
		Parameter: d.param,
	}, nil
}

func (d DummyParamGetter) GetParametersByPathWithContext(ctx context.Context, input *ssm.GetParametersByPathInput, opts ...request.Option) (*ssm.GetParametersByPathOutput, error) {
	if d.err != nil {
		return nil, d.err
	}
	assert.NotNil(d.t, d.params, "Must provide a param if no error!")
	return &ssm.GetParametersByPathOutput{
		Parameters: d.params,
	}, nil
}

func simpleAWSSourceHelper(dummy awssmpGetter) *Source {
	return &Source{
		Alias: "foo",
		URL: &url.URL{
			Scheme: "aws+smp",
			Path:   "/foo",
		},
		asmpg: dummy,
	}
}

func TestAWSSMP_ParseArgsSimple(t *testing.T) {
	u, _ := url.Parse("noddy")
	_, p, err := parseAWSSMPArgs(u)
	assert.Equal(t, "noddy", p)
	assert.Nil(t, err)
}

func TestAWSSMP_ParseArgsAppend(t *testing.T) {
	u, _ := url.Parse("base")
	_, p, err := parseAWSSMPArgs(u, "extra")
	assert.Equal(t, "base/extra", p)
	assert.Nil(t, err)
}

func TestAWSSMP_ParseArgsAppend2(t *testing.T) {
	u, _ := url.Parse("/foo/")
	_, p, err := parseAWSSMPArgs(u, "/extra")
	assert.Equal(t, "/foo/extra", p)
	assert.Nil(t, err)
}

func TestAWSSMP_ParseArgsTooMany(t *testing.T) {
	u, _ := url.Parse("base")
	_, _, err := parseAWSSMPArgs(u, "extra", "too many!")
	assert.Error(t, err)
}

func TestAWSSMP_GetParameterSetup(t *testing.T) {
	calledOk := false
	s := simpleAWSSourceHelper(DummyParamGetter{
		t: t,
		mockGetParameter: func(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
			assert.Equal(t, "/foo/bar", *input.Name)
			assert.True(t, *input.WithDecryption)
			calledOk = true
			return &ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{},
			}, nil
		},
	})

	_, err := readAWSSMP(s, "/bar")
	assert.True(t, calledOk)
	assert.Nil(t, err)
}

func TestAWSSMP_GetParameterValidOutput(t *testing.T) {
	expected := &ssm.Parameter{
		Name:    aws.String("/foo"),
		Type:    aws.String("String"),
		Value:   aws.String("val"),
		Version: aws.Int64(1),
	}
	s := simpleAWSSourceHelper(DummyParamGetter{
		t:     t,
		param: expected,
	})

	output, err := readAWSSMP(s, "")
	assert.Nil(t, err)
	actual := &ssm.Parameter{}
	err = json.Unmarshal(output, &actual)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
	assert.Equal(t, jsonMimetype, s.mediaType)
}

func TestAWSSMP_GetParameterMissing(t *testing.T) {
	expectedErr := awserr.New("ParameterNotFound", "Test of error message", nil)
	s := simpleAWSSourceHelper(DummyParamGetter{
		t:   t,
		err: expectedErr,
	})

	_, err := readAWSSMP(s, "")
	assert.Error(t, err, "Test of error message")
}

func TestAWSSMP_listAWSSMPParams(t *testing.T) {
	ctx := context.Background()
	s := simpleAWSSourceHelper(DummyParamGetter{
		t:   t,
		err: awserr.New("ParameterNotFound", "foo", nil),
	})
	_, err := listAWSSMPParams(ctx, s, "")
	assert.Error(t, err)

	s = simpleAWSSourceHelper(DummyParamGetter{
		t: t,
		params: []*ssm.Parameter{
			{Name: aws.String("/a")},
			{Name: aws.String("/b")},
			{Name: aws.String("/c")},
		},
	})
	data, err := listAWSSMPParams(ctx, s, "/")
	assert.NoError(t, err)
	assert.Equal(t, []byte(`["a","b","c"]`), data)

	s = simpleAWSSourceHelper(DummyParamGetter{
		t: t,
		params: []*ssm.Parameter{
			{Name: aws.String("/a/a")},
			{Name: aws.String("/a/b")},
			{Name: aws.String("/a/c")},
		},
	})
	data, err = listAWSSMPParams(ctx, s, "/a/")
	assert.NoError(t, err)
	assert.Equal(t, []byte(`["a","b","c"]`), data)
}
