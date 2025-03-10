package funcs

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateConvFuncs(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateConvFuncs(ctx)
			actual := fmap["conv"].(func() any)

			assert.Equal(t, ctx, actual().(*ConvFuncs).ctx)
		})
	}
}

func TestDefault(t *testing.T) {
	t.Parallel()

	s := struct{}{}
	c := &ConvFuncs{}
	def := "DEFAULT"
	data := []struct {
		val   any
		empty bool
	}{
		{0, true},
		{1, false},
		{nil, true},
		{"", true},
		{"foo", false},
		{[]string{}, true},
		{[]string{"foo"}, false},
		{[]string{""}, false},
		{c, false},
		{s, false},
	}

	for _, d := range data {
		t.Run(fmt.Sprintf("%T/%#v empty==%v", d.val, d.val, d.empty), func(t *testing.T) {
			t.Parallel()

			if d.empty {
				assert.Equal(t, def, c.Default(def, d.val))
			} else {
				assert.Equal(t, d.val, c.Default(def, d.val))
			}
		})
	}
}
