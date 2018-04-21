package funcs

import (
	"testing"

	"github.com/hairyhenderson/gomplate/aws"
	"github.com/stretchr/testify/assert"
)

func TestNSIsIdempotent(t *testing.T) {
	left := AWSNS()
	right := AWSNS()
	assert.True(t, left == right)
}
func TestFuncs(t *testing.T) {
	m := aws.NewDummyEc2Meta()
	i := aws.NewDummyEc2Info(m)
	af := &Funcs{meta: m, info: i}
	assert.Equal(t, "unknown", af.EC2Region())
	assert.Equal(t, "", af.EC2Meta("foo"))
	assert.Equal(t, "", af.EC2Tag("foo"))
	assert.Equal(t, "unknown", af.EC2Region())
}
