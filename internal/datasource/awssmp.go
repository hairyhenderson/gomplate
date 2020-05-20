package datasource

import (
	"bytes"
	"context"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"

	"github.com/ugorji/go/codec"

	gaws "github.com/hairyhenderson/gomplate/v3/aws"
)

// awssmpGetter - A subset of SSM API for use in unit testing
type awssmpGetter interface {
	GetParameterWithContext(ctx context.Context, input *ssm.GetParameterInput, opts ...request.Option) (*ssm.GetParameterOutput, error)
	GetParametersByPathWithContext(ctx context.Context, input *ssm.GetParametersByPathInput, opts ...request.Option) (*ssm.GetParametersByPathOutput, error)
}

// AWSSMP -
type AWSSMP struct {
	asmpg awssmpGetter
}

var _ Reader = (*AWSSMP)(nil)

func (a *AWSSMP) Read(ctx context.Context, url *url.URL, args ...string) (data *Data, err error) {
	if a.asmpg == nil {
		a.asmpg = ssm.New(gaws.SDKSession())
	}

	_, paramPath, err := parseAWSSMPArgs(url, args...)
	if err != nil {
		return nil, err
	}

	data = newData(url, args)

	data.MType = jsonMimetype
	switch {
	case strings.HasSuffix(paramPath, "/"):
		data.MType = jsonArrayMimetype
		data.Bytes, err = a.listAWSSMPParams(ctx, paramPath)
	default:
		data.Bytes, err = a.readAWSSMPParam(ctx, paramPath)
	}
	return data, err
}

func (a *AWSSMP) readAWSSMPParam(ctx context.Context, paramPath string) ([]byte, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(paramPath),
		WithDecryption: aws.Bool(true),
	}

	response, err := a.asmpg.GetParameterWithContext(ctx, input)

	if err != nil {
		return nil, errors.Wrapf(err, "Error reading aws+smp from AWS using GetParameter with input %v", input)
	}

	result := *response.Parameter

	output, err := toJSONBytes(result)
	return output, err
}

// listAWSSMPParams - supports directory semantics, returns array
func (a *AWSSMP) listAWSSMPParams(ctx context.Context, paramPath string) ([]byte, error) {
	input := &ssm.GetParametersByPathInput{
		Path: aws.String(paramPath),
	}

	response, err := a.asmpg.GetParametersByPathWithContext(ctx, input)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading aws+smp from AWS using GetParameter with input %v", input)
	}

	listing := make([]string, len(response.Parameters))
	for i, p := range response.Parameters {
		listing[i] = (*p.Name)[len(paramPath):]
	}

	output, err := toJSONBytes(listing)
	return output, err
}

func toJSONBytes(in interface{}) ([]byte, error) {
	h := &codec.JsonHandle{}
	h.Canonical = true
	buf := new(bytes.Buffer)
	err := codec.NewEncoder(buf, h).Encode(in)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to marshal %s", in)
	}
	return buf.Bytes(), nil
}

func parseAWSSMPArgs(sourceURL *url.URL, args ...string) (params map[string]interface{}, p string, err error) {
	if len(args) >= 2 {
		err = errors.New("Maximum two arguments to aws+smp datasource: alias, extraPath")
		return nil, "", err
	}

	p = sourceURL.Path
	params = make(map[string]interface{})
	for key, val := range sourceURL.Query() {
		params[key] = strings.Join(val, " ")
	}

	if len(args) == 1 {
		parsed, err := url.Parse(args[0])
		if err != nil {
			return nil, "", err
		}

		if parsed.Path != "" {
			p = path.Join(p, parsed.Path)
		}

		for key, val := range parsed.Query() {
			params[key] = strings.Join(val, " ")
		}
	}
	return params, p, err
}
