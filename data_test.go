package main

import (
	"net/url"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/stretchr/testify/assert"
)

func TestNewSource(t *testing.T) {
	s := NewSource("foo", &url.URL{
		Scheme: "file",
		Path:   "/foo.json",
	})
	assert.Equal(t, "application/json", s.Type)
	assert.Equal(t, ".json", s.Ext)

	s = NewSource("foo", &url.URL{
		Scheme: "http",
		Host:   "example.com",
		Path:   "/foo.json",
	})
	assert.Equal(t, "application/json", s.Type)
	assert.Equal(t, ".json", s.Ext)

	s = NewSource("foo", &url.URL{
		Scheme: "ftp",
		Host:   "example.com",
		Path:   "/foo.json",
	})
	assert.Equal(t, "application/json", s.Type)
	assert.Equal(t, ".json", s.Ext)
}

func TestParseSourceNoAlias(t *testing.T) {
	s, err := ParseSource("foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "foo", s.Alias)

	_, err = ParseSource("../foo.json")
	assert.Error(t, err)

	_, err = ParseSource("ftp://example.com/foo.yml")
	assert.Error(t, err)
}

func TestParseSourceWithAlias(t *testing.T) {
	s, err := ParseSource("data=foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.Equal(t, "application/json", s.Type)
	assert.True(t, s.URL.IsAbs())

	s, err = ParseSource("data=/otherdir/foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/otherdir/foo.json", s.URL.Path)

	s, err = ParseSource("data=sftp://example.com/blahblah/foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "sftp", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/blahblah/foo.json", s.URL.Path)
}

func TestDatasource(t *testing.T) {
	fs := memfs.Create()
	fs.Mkdir("/tmp", 0777)
	f, _ := vfs.Create(fs, "/tmp/foo.json")
	f.Write([]byte(`{"hello":"world"}`))

	sources := make(map[string]*Source)
	sources["foo"] = &Source{
		Alias: "foo",
		URL: &url.URL{
			Scheme: "file",
			Path:   "/tmp/foo.json",
		},
		Ext:  "json",
		Type: "application/json",
		FS:   fs,
	}
	data := &Data{
		Sources: sources,
	}
	expected := make(map[string]interface{})
	expected["hello"] = "world"
	actual := data.Datasource("foo")
	assert.Equal(t, expected["hello"], actual["hello"])
}
