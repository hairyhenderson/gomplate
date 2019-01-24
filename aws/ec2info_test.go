package aws

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

func TestTag_MissingKey(t *testing.T) {
	server, ec2meta := MockServer(200, `"i-1234"`)
	defer server.Close()
	client := DummyInstanceDescriber{
		tags: []*ec2.Tag{
			{
				Key:   aws.String("foo"),
				Value: aws.String("bar"),
			},
			{
				Key:   aws.String("baz"),
				Value: aws.String("qux"),
			},
		},
	}
	e := &Ec2Info{
		describer: func() (InstanceDescriber, error) {
			return client, nil
		},
		metaClient: ec2meta,
		cache:      make(map[string]interface{}),
	}

	assert.Empty(t, must(e.Tag("missing")))
	assert.Equal(t, "default", must(e.Tag("missing", "default")))
}

func TestTag_ValidKey(t *testing.T) {
	server, ec2meta := MockServer(200, `"i-1234"`)
	defer server.Close()
	client := DummyInstanceDescriber{
		tags: []*ec2.Tag{
			{
				Key:   aws.String("foo"),
				Value: aws.String("bar"),
			},
			{
				Key:   aws.String("baz"),
				Value: aws.String("qux"),
			},
		},
	}
	e := &Ec2Info{
		describer: func() (InstanceDescriber, error) {
			return client, nil
		},
		metaClient: ec2meta,
		cache:      make(map[string]interface{}),
	}

	assert.Equal(t, "bar", must(e.Tag("foo")))
	assert.Equal(t, "bar", must(e.Tag("foo", "default")))
}

func TestTag_NonEC2(t *testing.T) {
	server, ec2meta := MockServer(404, "")
	ec2meta.nonAWS = true
	defer server.Close()
	client := DummyInstanceDescriber{}
	e := &Ec2Info{
		describer: func() (InstanceDescriber, error) {
			return client, nil
		},
		metaClient: ec2meta,
		cache:      make(map[string]interface{}),
	}

	assert.Equal(t, "", must(e.Tag("foo")))
	assert.Equal(t, "default", must(e.Tag("foo", "default")))
}

func TestNewEc2Info(t *testing.T) {
	server, ec2meta := MockServer(200, `"i-1234"`)
	defer server.Close()
	client := DummyInstanceDescriber{
		tags: []*ec2.Tag{
			{
				Key:   aws.String("foo"),
				Value: aws.String("bar"),
			},
			{
				Key:   aws.String("baz"),
				Value: aws.String("qux"),
			},
		},
	}
	e := NewEc2Info(ClientOptions{})
	e.describer = func() (InstanceDescriber, error) {
		return client, nil
	}
	e.metaClient = ec2meta

	assert.Equal(t, "bar", must(e.Tag("foo")))
	assert.Equal(t, "bar", must(e.Tag("foo", "default")))
}

func TestGetRegion(t *testing.T) {
	oldReg, ok := os.LookupEnv("AWS_REGION")
	if ok {
		defer os.Setenv("AWS_REGION", oldReg)
	}
	oldDefReg, ok := os.LookupEnv("AWS_DEFAULT_REGION")
	if ok {
		defer os.Setenv("AWS_REGION", oldDefReg)
	}

	os.Setenv("AWS_REGION", "kalamazoo")
	os.Unsetenv("AWS_DEFAULT_REGION")
	region, err := getRegion()
	assert.NoError(t, err)
	assert.Empty(t, region)

	os.Setenv("AWS_DEFAULT_REGION", "kalamazoo")
	os.Unsetenv("AWS_REGION")
	region, err = getRegion()
	assert.NoError(t, err)
	assert.Empty(t, region)

	os.Unsetenv("AWS_DEFAULT_REGION")
	metaClient := NewDummyEc2Meta()
	region, err = getRegion(metaClient)
	assert.NoError(t, err)
	assert.Equal(t, "unknown", region)

	server, ec2meta := MockServer(200, `{"region":"us-east-1"}`)
	defer server.Close()
	region, err = getRegion(ec2meta)
	assert.NoError(t, err)
	assert.Equal(t, "us-east-1", region)
}

func TestGetClientOptions(t *testing.T) {
	oldVar, ok := os.LookupEnv("AWS_TIMEOUT")
	if ok {
		defer os.Setenv("AWS_TIMEOUT", oldVar)
	}

	co := GetClientOptions()
	assert.Equal(t, ClientOptions{Timeout: 500 * time.Millisecond}, co)

	os.Setenv("AWS_TIMEOUT", "42")
	// reset the Once
	coInit = sync.Once{}
	co = GetClientOptions()
	assert.Equal(t, ClientOptions{Timeout: 42 * time.Millisecond}, co)

	os.Setenv("AWS_TIMEOUT", "123")
	// without resetting the Once, expect to be reused
	co = GetClientOptions()
	assert.Equal(t, ClientOptions{Timeout: 42 * time.Millisecond}, co)

	os.Setenv("AWS_TIMEOUT", "foo")
	// reset the Once
	coInit = sync.Once{}
	assert.Panics(t, func() {
		GetClientOptions()
	})
}
