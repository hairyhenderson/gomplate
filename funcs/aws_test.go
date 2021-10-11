package funcs

import (
	"context"
	"strconv"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/aws"
	"github.com/stretchr/testify/assert"
)

func TestCreateAWSFuncs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateAWSFuncs(ctx)
			actual := fmap["aws"].(func() interface{})

			assert.Same(t, ctx, actual().(*Funcs).ctx)
		})
	}
}

func TestAWSFuncs(t *testing.T) {
	t.Parallel()

	m := aws.NewDummyEc2Meta()
	i := aws.NewDummyEc2Info(m)
	af := &Funcs{meta: m, info: i}
	assert.Equal(t, "unknown", must(af.EC2Region()))
	assert.Equal(t, "", must(af.EC2Meta("foo")))
	assert.Equal(t, "", must(af.EC2Tag("foo")))
	assert.Equal(t, "unknown", must(af.EC2Region()))
}

func must(r interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return r
}
