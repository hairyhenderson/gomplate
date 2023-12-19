package aws

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTag_MissingKey(t *testing.T) {
	ec2meta := MockEC2Meta(map[string]string{"instance-id": "i-1234"}, nil, "")

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
	ec2meta := MockEC2Meta(map[string]string{"instance-id": "i-1234"}, nil, "")

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

func TestTags(t *testing.T) {
	ec2meta := MockEC2Meta(map[string]string{"instance-id": "i-1234"}, nil, "")
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

	assert.Equal(t, map[string]string{"foo": "bar", "baz": "qux"}, must(e.Tags()))
}

func TestTag_NonEC2(t *testing.T) {
	ec2meta := MockEC2Meta(nil, nil, "")
	ec2meta.nonAWS = true

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
	ec2meta := MockEC2Meta(map[string]string{"instance-id": "i-1234"}, nil, "")
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
	// unset AWS region env vars for clean tests
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_DEFAULT_REGION")

	t.Run("with AWS_REGION set", func(t *testing.T) {
		t.Setenv("AWS_REGION", "kalamazoo")
		region, err := getRegion()
		require.NoError(t, err)
		assert.Empty(t, region)
	})

	t.Run("with AWS_DEFAULT_REGION set", func(t *testing.T) {
		t.Setenv("AWS_DEFAULT_REGION", "kalamazoo")
		region, err := getRegion()
		require.NoError(t, err)
		assert.Empty(t, region)
	})

	t.Run("with no AWS_REGION, AWS_DEFAULT_REGION set", func(t *testing.T) {
		metaClient := NewDummyEc2Meta()
		region, err := getRegion(metaClient)
		require.NoError(t, err)
		assert.Equal(t, "unknown", region)
	})

	t.Run("infer from EC2 metadata", func(t *testing.T) {
		ec2meta := MockEC2Meta(nil, nil, "us-east-1")
		region, err := getRegion(ec2meta)
		require.NoError(t, err)
		assert.Equal(t, "us-east-1", region)
	})
}

func TestGetClientOptions(t *testing.T) {
	co := GetClientOptions()
	assert.Equal(t, ClientOptions{Timeout: 500 * time.Millisecond}, co)

	t.Run("valid AWS_TIMEOUT, first call", func(t *testing.T) {
		t.Setenv("AWS_TIMEOUT", "42")
		// reset the Once
		coInit = sync.Once{}
		co = GetClientOptions()
		assert.Equal(t, ClientOptions{Timeout: 42 * time.Millisecond}, co)
	})

	t.Run("valid AWS_TIMEOUT, non-first call", func(t *testing.T) {
		t.Setenv("AWS_TIMEOUT", "123")
		// without resetting the Once, expect to be reused
		co = GetClientOptions()
		assert.Equal(t, ClientOptions{Timeout: 42 * time.Millisecond}, co)
	})

	t.Run("invalid AWS_TIMEOUT", func(t *testing.T) {
		t.Setenv("AWS_TIMEOUT", "foo")
		// reset the Once
		coInit = sync.Once{}
		assert.Panics(t, func() {
			GetClientOptions()
		})
	})
}
