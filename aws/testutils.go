package aws

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// MockServer -
func MockServer(code int, body string) (*httptest.Server, *Ec2Meta) {
	server, httpClient := MockHTTPServer(code, body)

	client := &Ec2Meta{server.URL + "/", httpClient, false, make(map[string]string), ClientOptions{}}
	return server, client
}

// MockHTTPServer -
func MockHTTPServer(code int, body string) (*httptest.Server, *http.Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		// nolint: errcheck
		fmt.Fprintln(w, body)
	}))

	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}
	httpClient := &http.Client{Transport: tr}
	return server, httpClient
}

// NewDummyEc2Info -
func NewDummyEc2Info(metaClient *Ec2Meta) *Ec2Info {
	i := &Ec2Info{
		metaClient: metaClient,
		describer:  func() (InstanceDescriber, error) { return DummyInstanceDescriber{}, nil },
	}
	return i
}

// NewDummyEc2Meta -
func NewDummyEc2Meta() *Ec2Meta {
	return &Ec2Meta{nonAWS: true}
}

// DummyInstanceDescriber - test doubles
type DummyInstanceDescriber struct {
	tags []*ec2.Tag
}

// DescribeInstances -
func (d DummyInstanceDescriber) DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	output := &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
					{
						Tags: d.tags,
					},
				},
			},
		},
	}
	return output, nil
}
