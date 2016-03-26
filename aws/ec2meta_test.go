package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEc2meta_MissingKey(t *testing.T) {
	server, ec2meta := MockServer(404, "")
	defer server.Close()

	assert.Empty(t, ec2meta.Ec2meta("foo"))
	assert.Equal(t, "default", ec2meta.Ec2meta("foo", "default"))
}

func TestEc2meta_ValidKey(t *testing.T) {
	server, ec2meta := MockServer(200, "i-1234")
	defer server.Close()

	assert.Equal(t, "i-1234", ec2meta.Ec2meta("instance-id"))
	assert.Equal(t, "i-1234", ec2meta.Ec2meta("instance-id", "unused default"))
}

func TestEc2dynamic_MissingKey(t *testing.T) {
	server, ec2meta := MockServer(404, "")
	defer server.Close()

	assert.Empty(t, ec2meta.Ec2dynamic("foo"))
	assert.Equal(t, "default", ec2meta.Ec2dynamic("foo", "default"))
}

func TestEc2dynamic_ValidKey(t *testing.T) {
	server, ec2meta := MockServer(200, "i-1234")
	defer server.Close()

	assert.Equal(t, "i-1234", ec2meta.Ec2dynamic("instance-id"))
	assert.Equal(t, "i-1234", ec2meta.Ec2dynamic("instance-id", "unused default"))
}
