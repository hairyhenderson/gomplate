package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

// test doubles
type DummyInstanceDescriber struct {
	tags []*ec2.Tag
}

func (d DummyInstanceDescriber) DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	output := &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			&ec2.Reservation{
				Instances: []*ec2.Instance{
					&ec2.Instance{
						Tags: d.tags,
					},
				},
			},
		},
	}
	return output, nil
}

func TestTag_MissingKey(t *testing.T) {
	server, ec2meta := MockServer(200, `"i-1234"`)
	defer server.Close()
	client := DummyInstanceDescriber{
		tags: []*ec2.Tag{
			&ec2.Tag{
				Key:   aws.String("foo"),
				Value: aws.String("bar"),
			},
			&ec2.Tag{
				Key:   aws.String("baz"),
				Value: aws.String("qux"),
			},
		},
	}
	e := &Ec2Info{
		describer: func() InstanceDescriber {
			return client
		},
		metaClient: ec2meta,
		cache:      make(map[string]interface{}),
	}

	assert.Empty(t, e.Tag("missing"))
	assert.Equal(t, "default", e.Tag("missing", "default"))
}

func TestTag_ValidKey(t *testing.T) {
	server, ec2meta := MockServer(200, `"i-1234"`)
	defer server.Close()
	client := DummyInstanceDescriber{
		tags: []*ec2.Tag{
			&ec2.Tag{
				Key:   aws.String("foo"),
				Value: aws.String("bar"),
			},
			&ec2.Tag{
				Key:   aws.String("baz"),
				Value: aws.String("qux"),
			},
		},
	}
	e := &Ec2Info{
		describer: func() InstanceDescriber {
			return client
		},
		metaClient: ec2meta,
		cache:      make(map[string]interface{}),
	}

	assert.Equal(t, "bar", e.Tag("foo"))
	assert.Equal(t, "bar", e.Tag("foo", "default"))
}

func TestTag_NonEC2(t *testing.T) {
	server, ec2meta := MockServer(404, "")
	ec2meta.nonAWS = true
	defer server.Close()
	client := DummyInstanceDescriber{}
	e := &Ec2Info{
		describer: func() InstanceDescriber {
			return client
		},
		metaClient: ec2meta,
		cache:      make(map[string]interface{}),
	}

	assert.Equal(t, "", e.Tag("foo"))
	assert.Equal(t, "default", e.Tag("foo", "default"))
}
