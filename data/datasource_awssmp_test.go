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
	err              awserr.Error
	t                *testing.T
	param            *ssm.Parameter
	mockGetParameter func(*ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
	params           []*ssm.Parameter
}

func (d DummyParamGetter) GetParameterWithContext(_ context.Context, input *ssm.GetParameterInput, _ ...request.Option) (*ssm.GetParameterOutput, error) {
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

func (d DummyParamGetter) GetParametersByPathWithContext(_ context.Context, _ *ssm.GetParametersByPathInput, _ ...request.Option) (*ssm.GetParametersByPathOutput, error) {
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

	_, err := readAWSSMP(context.Background(), s, "/bar")
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

	output, err := readAWSSMP(context.Background(), s, "")
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

	_, err := readAWSSMP(context.Background(), s, "")
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
