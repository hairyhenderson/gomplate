// +build !windows

package data

import (
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/stretchr/testify/assert"
)

// DummyParamGetter - test double
type DummyParamGetter struct {
	t                *testing.T
	param            *ssm.Parameter
	err              awserr.Error
	mockGetParameter func(*ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
}

func (d DummyParamGetter) GetParameter(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
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

func simpleAWSSourceHelper(dummy AWSSMPGetter) *Source {
	return &Source{
		Alias: "foo",
		URL: &url.URL{
			Scheme: "aws+smp",
			Path:   "/foo",
		},
		ASMPG: dummy,
	}
}

func TestAWSSMP_ParseArgsSimple(t *testing.T) {
	paramPath, err := parseAWSSMPArgs("noddy")
	assert.Equal(t, "noddy", paramPath)
	assert.Nil(t, err)
}

func TestAWSSMP_ParseArgsAppend(t *testing.T) {
	paramPath, err := parseAWSSMPArgs("base", "extra")
	assert.Equal(t, "base/extra", paramPath)
	assert.Nil(t, err)
}

func TestAWSSMP_ParseArgsAppend2(t *testing.T) {
	paramPath, err := parseAWSSMPArgs("/foo/", "/extra")
	assert.Equal(t, "/foo/extra", paramPath)
	assert.Nil(t, err)
}

func TestAWSSMP_ParseArgsTooMany(t *testing.T) {
	_, err := parseAWSSMPArgs("base", "extra", "too many!")
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
	s := simpleAWSSourceHelper(DummyParamGetter{
		t: t,
		param: &ssm.Parameter{
			Name:    aws.String("/foo"),
			Type:    aws.String("String"),
			Value:   aws.String("val"),
			Version: aws.Int64(1),
		},
	})

	output, err := readAWSSMP(s, "")
	assert.Nil(t, err)
	expected := "{\"Name\":\"/foo\",\"Type\":\"String\",\"Value\":\"val\",\"Version\":1}"
	assert.Equal(t, []byte(expected), output)
	assert.Equal(t, jsonMimetype, s.Type)
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
