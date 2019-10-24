package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessIncludes(t *testing.T) {
	data := []struct {
		inc, exc, expected []string
	}{
		{nil, nil, []string{}},
		{[]string{}, []string{}, []string{}},
		{nil, []string{"*.foo"}, []string{"*.foo"}},
		{[]string{"*.bar"}, []string{"a*.bar"}, []string{"*", "!*.bar", "a*.bar"}},
		{[]string{"*.bar"}, nil, []string{"*", "!*.bar"}},
	}

	for _, d := range data {
		assert.EqualValues(t, d.expected, processIncludes(d.inc, d.exc))
	}
}
