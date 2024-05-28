package datafs

import (
	"net/http"
	"testing"

	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/stretchr/testify/require"
)

func TestDefaultRegistry(t *testing.T) {
	reg := NewRegistry()
	ds := config.DataSource{}
	reg.Register("foo", ds)

	actual, ok := reg.Lookup("foo")
	require.True(t, ok)
	require.Equal(t, ds, actual)

	_, ok = reg.Lookup("bar")
	require.False(t, ok)
}

func TestDefaultRegistry_List(t *testing.T) {
	reg := NewRegistry()
	ds := config.DataSource{}
	reg.Register("a", ds)
	reg.Register("b", ds)
	reg.Register("c", ds)
	reg.Register("d", ds)

	actual := reg.List()

	// list must be sorted
	require.Equal(t, []string{"a", "b", "c", "d"}, actual)
}

func TestDefaultRegistry_AddExtraHeader(t *testing.T) {
	reg := NewRegistry()
	hdr := http.Header{"foo": {"bar"}}
	reg.AddExtraHeader("baz", hdr)

	reg.Register("baz", config.DataSource{})

	ds, ok := reg.Lookup("baz")
	require.True(t, ok)
	require.Equal(t, hdr, ds.Header)
}
