package datasources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	gaws "github.com/hairyhenderson/gomplate/v3/aws"
)

type awsSMPRequester struct {
	asmpg awssmpGetter
}

// awssmpGetter - A subset of SSM API for use in unit testing
type awssmpGetter interface {
	GetParameterWithContext(ctx context.Context, input *ssm.GetParameterInput, opts ...request.Option) (*ssm.GetParameterOutput, error)
	GetParametersByPathWithContext(ctx context.Context, input *ssm.GetParametersByPathInput, opts ...request.Option) (*ssm.GetParametersByPathOutput, error)
}

func (r *awsSMPRequester) Request(ctx context.Context, u *url.URL, header http.Header) (*Response, error) {
	if r.asmpg == nil {
		r.asmpg = ssm.New(gaws.SDKSession())
	}

	resp := &Response{}
	ct := jsonMimetype

	var err error
	var out interface{}
	switch {
	case strings.HasSuffix(u.Path, "/"):
		ct = jsonArrayMimetype
		out, err = r.listAWSSMPParams(ctx, u.Path)
	default:
		out, err = r.readAWSSMPParam(ctx, u.Path)
	}
	if err != nil {
		return nil, err
	}

	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	err = enc.Encode(out)
	if err != nil {
		return nil, err
	}

	ct, _ = mimeType(u, ct)
	resp.ContentType = ct
	resp.Body = ioutil.NopCloser(b)
	resp.ContentLength = int64(b.Len())

	return resp, err
}

func (r *awsSMPRequester) readAWSSMPParam(ctx context.Context, paramPath string) (*ssm.Parameter, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(paramPath),
		WithDecryption: aws.Bool(true),
	}

	response, err := r.asmpg.GetParameterWithContext(ctx, input)

	if err != nil {
		return nil, fmt.Errorf("failed to read aws+smp from AWS using GetParameterWithContext with input %v: %w", input, err)
	}

	return response.Parameter, nil
}

// listAWSSMPParams - supports directory semantics, returns array
func (r *awsSMPRequester) listAWSSMPParams(ctx context.Context, paramPath string) ([]string, error) {
	input := &ssm.GetParametersByPathInput{
		Path: aws.String(paramPath),
	}

	response, err := r.asmpg.GetParametersByPathWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to read aws+smp from AWS using GetParametersByPathWithContext with input %v: %w", input, err)
	}

	listing := make([]string, len(response.Parameters))
	for i, p := range response.Parameters {
		listing[i] = (*p.Name)[len(paramPath):]
	}
	return listing, nil
}
