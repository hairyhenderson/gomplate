package datasources

import (
	"context"
	"encoding/json"
	"io/ioutil"
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
	err              awserr.Error
	mockGetParameter func(*ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
	params           []*ssm.Parameter
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

func TestAWSSMPRequest(t *testing.T) {
	expectedErr := awserr.New("ParameterNotFound", "Test of error message", nil)
	ctx := context.Background()
	r := &awsSMPRequester{DummyParamGetter{
		t:   t,
		err: expectedErr,
	}}

	_, err := r.Request(ctx, mustParseURL("aws+smp:///foo"), nil)
	assert.Error(t, err, "Test of error message")

	r = &awsSMPRequester{DummyParamGetter{
		t: t,
		mockGetParameter: func(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
			assert.Equal(t, "/foo/bar", *input.Name)
			assert.True(t, *input.WithDecryption)
			return &ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{},
			}, nil
		},
	}}

	_, err = r.Request(ctx, mustParseURL("aws+smp:///foo/bar"), nil)
	assert.Nil(t, err)
}

func TestAWSSMP_GetParameterValidOutput(t *testing.T) {
	expected := &ssm.Parameter{
		Name:    aws.String("/foo"),
		Type:    aws.String("String"),
		Value:   aws.String("val"),
		Version: aws.Int64(1),
	}
	ctx := context.Background()
	r := &awsSMPRequester{DummyParamGetter{
		t:     t,
		param: expected,
	}}

	resp, err := r.Request(ctx, mustParseURL("aws+smp:///foo"), nil)
	assert.Nil(t, err)
	b, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	actual := &ssm.Parameter{}
	err = json.Unmarshal(b, actual)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)
	assert.Equal(t, jsonMimetype, resp.ContentType)
}

func TestAWSSMP_listAWSSMPParams(t *testing.T) {
	ctx := context.Background()
	r := &awsSMPRequester{DummyParamGetter{
		t:   t,
		err: awserr.New("ParameterNotFound", "foo", nil),
	}}
	_, err := r.listAWSSMPParams(ctx, "")
	assert.Error(t, err)

	r = &awsSMPRequester{DummyParamGetter{
		t: t,
		params: []*ssm.Parameter{
			{Name: aws.String("/a")},
			{Name: aws.String("/b")},
			{Name: aws.String("/c")},
		},
	}}

	out, err := r.listAWSSMPParams(ctx, "/")
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, out)

	r = &awsSMPRequester{DummyParamGetter{
		t: t,
		params: []*ssm.Parameter{
			{Name: aws.String("/a/a")},
			{Name: aws.String("/a/b")},
			{Name: aws.String("/a/c")},
		},
	}}
	out, err = r.listAWSSMPParams(ctx, "/a/")
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, out)
}
