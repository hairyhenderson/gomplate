package aws

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func must(r any, err error) any {
	if err != nil {
		panic(err)
	}
	return r
}

func TestMeta_MissingKey(t *testing.T) {
	ec2meta := MockEC2Meta(nil, nil, "")

	assert.Empty(t, must(ec2meta.Meta(t.Context(), "foo")))
	assert.Equal(t, "default", must(ec2meta.Meta(t.Context(), "foo", "default")))
}

func TestMeta_ValidKey(t *testing.T) {
	ec2meta := MockEC2Meta(map[string]string{"instance-id": "i-1234"}, nil, "")

	assert.Equal(t, "i-1234", must(ec2meta.Meta(t.Context(), "instance-id")))
	assert.Equal(t, "i-1234", must(ec2meta.Meta(t.Context(), "instance-id", "unused default")))
}

func TestDynamic_MissingKey(t *testing.T) {
	ec2meta := MockEC2Meta(nil, nil, "")

	assert.Empty(t, must(ec2meta.Dynamic(t.Context(), "foo")))
	assert.Equal(t, "default", must(ec2meta.Dynamic(t.Context(), "foo", "default")))
}

func TestDynamic_ValidKey(t *testing.T) {
	ec2meta := MockEC2Meta(nil, map[string]string{"instance-id": "i-1234"}, "")

	assert.Equal(t, "i-1234", must(ec2meta.Dynamic(t.Context(), "instance-id")))
	assert.Equal(t, "i-1234", must(ec2meta.Dynamic(t.Context(), "instance-id", "unused default")))
}

func TestRegion_NoRegion(t *testing.T) {
	ec2meta := MockEC2Meta(nil, nil, "")

	assert.Equal(t, "unknown", must(ec2meta.Region(t.Context())))
}

func TestRegion_NoRegionWithDefault(t *testing.T) {
	ec2meta := MockEC2Meta(nil, nil, "")

	assert.Equal(t, "foo", must(ec2meta.Region(t.Context(), "foo")))
}

func TestRegion_KnownRegion(t *testing.T) {
	ec2meta := MockEC2Meta(nil, nil, "us-east-1")

	assert.Equal(t, "us-east-1", must(ec2meta.Region(t.Context())))
}

func TestUnreachable(t *testing.T) {
	assert.False(t, unreachable(errors.New("foo")))
	assert.True(t, unreachable(errors.New("host is down")))
	assert.True(t, unreachable(errors.New("request canceled")))
	assert.True(t, unreachable(errors.New("no route to host")))
}

func TestRetrieveMetadata_NonEC2(t *testing.T) {
	ec2meta := NewEc2Meta()
	ec2meta.nonAWS = true

	assert.Equal(t, "foo", must(ec2meta.retrieveMetadata(t.Context(), "", "foo")))
}
