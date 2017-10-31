package time

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestZoneFuncs(t *testing.T) {
	name, offset := time.Now().Zone()
	assert.Equal(t, name, ZoneName())
	assert.Equal(t, offset, ZoneOffset())
}
