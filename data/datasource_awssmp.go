package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"

	gaws "github.com/hairyhenderson/gomplate/v3/aws"
)

// awssmpGetter - A subset of SSM API for use in unit testing
type awssmpGetter interface {
	GetParameterWithContext(ctx context.Context, input *ssm.GetParameterInput, opts ...request.Option) (*ssm.GetParameterOutput, error)
	GetParametersByPathWithContext(ctx context.Context, input *ssm.GetParametersByPathInput, opts ...request.Option) (*ssm.GetParametersByPathOutput, error)
}

func readAWSSMP(ctx context.Context, source *Source, args ...string) (data []byte, err error) {
	if source.asmpg == nil {
		source.asmpg = ssm.New(gaws.SDKSession())
	}

	_, paramPath, err := parseDatasourceURLArgs(source.URL, args...)
	if err != nil {
		return nil, err
	}

	source.mediaType = jsonMimetype
	switch {
	case strings.HasSuffix(paramPath, "/"):
		source.mediaType = jsonArrayMimetype
		data, err = listAWSSMPParams(ctx, source, paramPath)
	default:
		data, err = readAWSSMPParam(ctx, source, paramPath)
	}
	return data, err
}

func readAWSSMPParam(ctx context.Context, source *Source, paramPath string) ([]byte, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(paramPath),
		WithDecryption: aws.Bool(true),
	}

	response, err := source.asmpg.GetParameterWithContext(ctx, input)

	if err != nil {
		return nil, fmt.Errorf("error reading aws+smp from AWS using GetParameter with input %v: %w", input, err)
	}

	result := *response.Parameter

	output, err := ToJSON(result)
	return []byte(output), err
}

// listAWSSMPParams - supports directory semantics, returns array
func listAWSSMPParams(ctx context.Context, source *Source, paramPath string) ([]byte, error) {
	input := &ssm.GetParametersByPathInput{
		Path: aws.String(paramPath),
	}

	response, err := source.asmpg.GetParametersByPathWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error reading aws+smp from AWS using GetParameter with input %v: %w", input, err)
	}

	listing := make([]string, len(response.Parameters))
	for i, p := range response.Parameters {
		listing[i] = (*p.Name)[len(paramPath):]
	}

	output, err := ToJSON(listing)
	return []byte(output), err
}
