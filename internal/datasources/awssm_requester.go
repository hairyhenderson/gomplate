package datasources

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	gaws "github.com/hairyhenderson/gomplate/v3/aws"
)

type awsSecretsManagerRequester struct {
	awsSecretsManager awsSecretsManagerGetter
}

func (r *awsSecretsManagerRequester) Request(ctx context.Context, u *url.URL, header http.Header) (*Response, error) {
	if r.awsSecretsManager == nil {
		r.awsSecretsManager = secretsmanager.New(gaws.SDKSession())
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(u.Path),
	}

	response, err := r.awsSecretsManager.GetSecretValueWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to read from AWS using GetSecretValue with input %v: %w", input, err)
	}

	ct, err := mimeType(u, "")
	if err != nil {
		ct = textMimetype
	}
	resp := &Response{
		Body:          ioutil.NopCloser(bytes.NewBufferString(*response.SecretString)),
		ContentLength: int64(len(*response.SecretString)),
		ContentType:   ct,
	}

	return resp, nil
}

// awsSecretsManagerGetter - A subset of Secrets Manager API for use in unit testing
type awsSecretsManagerGetter interface {
	GetSecretValueWithContext(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error)
}
