package funcs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	s := struct{}{}
	c := ConvNS()
	def := "DEFAULT"
	data := []struct {
		val   interface{}
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
		d := d
		t.Run(fmt.Sprintf("%T/%#v empty==%v", d.val, d.val, d.empty), func(t *testing.T) {
			if d.empty {
				assert.Equal(t, def, c.Default(def, d.val))
			} else {
				assert.Equal(t, d.val, c.Default(def, d.val))
			}
		})
	}
}
