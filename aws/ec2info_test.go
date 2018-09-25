package aws

import (
	"testing"

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
