package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetLookupIP(t *testing.T) {
	n := &NetFuncs{}
	assert.Equal(t, "127.0.0.1", must(n.LookupIP("localhost")))
}
