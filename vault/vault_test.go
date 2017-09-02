package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	server, v := MockServer(404, "Not Found")
	defer server.Close()
	val, err := v.Read("secret/bogus")
	assert.Empty(t, val)
	assert.NoError(t, err)

	expected := "{\"value\":\"foo\"}\n"
	server, v = MockServer(200, `{"data":`+expected+`}`)
	defer server.Close()
	val, err = v.Read("s")
	assert.Equal(t, expected, string(val))
	assert.NoError(t, err)
}

func TestWrite(t *testing.T) {
	server, v := MockServer(404, "Not Found")
	defer server.Close()
	val, err := v.Write("secret/bogus", nil)
	assert.Empty(t, val)
	assert.Error(t, err)

	expected := "{\"value\":\"foo\"}\n"
	server, v = MockServer(200, `{"data":`+expected+`}`)
	defer server.Close()
	val, err = v.Write("s", nil)
	assert.Equal(t, expected, string(val))
	assert.NoError(t, err)
}
