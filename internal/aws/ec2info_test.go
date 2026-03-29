package aws

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

func TestTag_MissingKey(t *testing.T) {
	ec2meta := MockEC2Meta(map[string]string{"instance-id": "i-1234"}, nil, "")

	client := DummyInstanceDescriber{
		tags: []types.Tag{
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
		describer: func(_ context.Context) (InstanceDescriber, error) {
			return client, nil
		},
		metaClient: ec2meta,
		cache:      make(map[string]any),
	}

	assert.Empty(t, must(e.Tag(t.Context(), "missing")))
	assert.Equal(t, "default", must(e.Tag(t.Context(), "missing", "default")))
}

func TestTag_ValidKey(t *testing.T) {
	ec2meta := MockEC2Meta(map[string]string{"instance-id": "i-1234"}, nil, "")

	client := DummyInstanceDescriber{
		tags: []types.Tag{
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
		describer: func(_ context.Context) (InstanceDescriber, error) {
			return client, nil
		},
		metaClient: ec2meta,
		cache:      make(map[string]any),
	}

	assert.Equal(t, "bar", must(e.Tag(t.Context(), "foo")))
	assert.Equal(t, "bar", must(e.Tag(t.Context(), "foo", "default")))
}

func TestTags(t *testing.T) {
	ec2meta := MockEC2Meta(map[string]string{"instance-id": "i-1234"}, nil, "")
	client := DummyInstanceDescriber{
		tags: []types.Tag{
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
		describer: func(_ context.Context) (InstanceDescriber, error) {
			return client, nil
		},
		metaClient: ec2meta,
		cache:      make(map[string]any),
	}

	assert.Equal(t, map[string]string{"foo": "bar", "baz": "qux"}, must(e.Tags(t.Context())))
}

func TestTag_NonEC2(t *testing.T) {
	ec2meta := MockEC2Meta(nil, nil, "")
	ec2meta.nonAWS = true

	client := DummyInstanceDescriber{}
	e := &Ec2Info{
		describer: func(_ context.Context) (InstanceDescriber, error) {
			return client, nil
		},
		metaClient: ec2meta,
		cache:      make(map[string]any),
	}

	assert.Empty(t, must(e.Tag(t.Context(), "foo")))
	assert.Equal(t, "default", must(e.Tag(t.Context(), "foo", "default")))
}

func TestNewEc2Info(t *testing.T) {
	ec2meta := MockEC2Meta(map[string]string{"instance-id": "i-1234"}, nil, "")
	client := DummyInstanceDescriber{
		tags: []types.Tag{
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
	e := NewEc2Info()
	e.describer = func(_ context.Context) (InstanceDescriber, error) {
		return client, nil
	}
	e.metaClient = ec2meta

	assert.Equal(t, "bar", must(e.Tag(t.Context(), "foo")))
	assert.Equal(t, "bar", must(e.Tag(t.Context(), "foo", "default")))
}

func TestGetClientOptions(t *testing.T) {
	co := getClientOptions()
	assert.Equal(t, ClientOptions{Timeout: 500 * time.Millisecond}, co)

	t.Run("valid AWS_TIMEOUT, first call", func(t *testing.T) {
		t.Setenv("AWS_TIMEOUT", "42")
		// reset the Once
		coInit = sync.Once{}
		co = getClientOptions()
		assert.Equal(t, ClientOptions{Timeout: 42 * time.Millisecond}, co)
	})

	t.Run("valid AWS_TIMEOUT, non-first call", func(t *testing.T) {
		t.Setenv("AWS_TIMEOUT", "123")
		// without resetting the Once, expect to be reused
		co = getClientOptions()
		assert.Equal(t, ClientOptions{Timeout: 42 * time.Millisecond}, co)
	})

	t.Run("invalid AWS_TIMEOUT", func(t *testing.T) {
		t.Setenv("AWS_TIMEOUT", "foo")
		// reset the Once
		coInit = sync.Once{}
		assert.Panics(t, func() {
			getClientOptions()
		})
	})
}
