package funcs

import (
	"context"
	"strconv"
	"testing"

	"github.com/hairyhenderson/gomplate/v4/aws"
	"github.com/stretchr/testify/assert"
)

func TestCreateAWSFuncs(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateAWSFuncs(ctx)
			actual := fmap["aws"].(func() any)

			assert.Equal(t, ctx, actual().(*Funcs).ctx)
		})
	}
}

func TestAWSFuncs(t *testing.T) {
	t.Parallel()

	m := aws.NewDummyEc2Meta()
	i := aws.NewDummyEc2Info(m)
	af := &Funcs{meta: m, info: i}
	assert.Equal(t, "unknown", must(af.EC2Region()))
	assert.Empty(t, must(af.EC2Meta("foo")))
	assert.Empty(t, must(af.EC2Tag("foo")))
	assert.Equal(t, "unknown", must(af.EC2Region()))
}

func must(r any, err error) any {
	if err != nil {
		panic(err)
	}
	return r
}
