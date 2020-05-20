package datasource

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheKey(t *testing.T) {
	s := &src{}
	assert.Equal(t, "", s.cacheKey())

	s = &src{alias: "foo"}
	assert.Equal(t, "foo", s.cacheKey())

	s = &src{alias: "bar"}
	assert.Equal(t, "barf", s.cacheKey("f"))

	s = &src{alias: "bar"}
	assert.Equal(t, "barfbag", s.cacheKey("f", "ba", "g"))
}

func TestNewSource(t *testing.T) {
	r := &srcRegistry{}
	s, err := r.Register("", mustParseURL("bogus:///"), nil)
	assert.Error(t, err)
	assert.Nil(t, s)
}

type dummyReader struct {
	err  error
	data *Data

	url  *url.URL
	args []string
}

func (d *dummyReader) Read(ctx context.Context, url *url.URL, args ...string) (*Data, error) {
	d.url = url
	d.args = args
	return d.data, d.err
}

func TestRead(t *testing.T) {
	u := mustParseURL("http://example.com")
	r := &dummyReader{
		err: fmt.Errorf("error"),
	}
	s := &src{alias: "foo", url: u, r: r}
	ctx := context.Background()

	d, err := s.Read(ctx)
	assert.Error(t, err)
	assert.Nil(t, d)
	assert.Equal(t, u, r.url)

	expected := &Data{Bytes: []byte("hello world")}
	r = &dummyReader{
		data: expected,
	}
	s = &src{alias: "foo", url: u, r: r}
	d, err = s.Read(ctx)
	assert.NoError(t, err)
	assert.Equal(t, u, r.url)
	assert.Equal(t, expected, d)

	// this data should not be used, instead cached data from previous call
	// should be returned
	r = &dummyReader{
		data: &Data{Bytes: []byte("goodbye world")},
	}
	s = &src{alias: "foo", url: u, r: r}
	d, err = s.Read(ctx)
	assert.NoError(t, err)
	assert.Nil(t, r.url)
	assert.Equal(t, expected, d)

	// this data should be used, previous cache was with no args
	expected = &Data{Bytes: []byte("goodbye world")}
	r = &dummyReader{
		data: expected,
	}
	s = &src{alias: "foo", url: u, r: r}
	d, err = s.Read(ctx, "bar")
	assert.NoError(t, err)
	assert.Equal(t, u, r.url)
	assert.EqualValues(t, []string{"bar"}, r.args)
	assert.Equal(t, expected, d)
}
