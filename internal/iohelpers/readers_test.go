package iohelpers

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLazyReadCloser(t *testing.T) {
	r := newBufferCloser(bytes.NewBufferString("hello world"))
	opened := false
	l, ok := LazyReadCloser(func() (io.ReadCloser, error) {
		opened = true
		return r, nil
	}).(*lazyReadCloser)
	assert.True(t, ok)

	assert.False(t, opened)
	assert.Nil(t, l.r)
	assert.False(t, r.closed)

	p := make([]byte, 5)
	n, err := l.Read(p)
	require.NoError(t, err)
	assert.True(t, opened)
	assert.Equal(t, r, l.r)
	assert.Equal(t, 5, n)

	err = l.Close()
	require.NoError(t, err)
	assert.True(t, r.closed)

	// test error propagation
	l = LazyReadCloser(func() (io.ReadCloser, error) {
		return nil, os.ErrNotExist
	}).(*lazyReadCloser)

	assert.Nil(t, l.r)

	p = make([]byte, 5)
	_, err = l.Read(p)
	require.Error(t, err)

	err = l.Close()
	require.Error(t, err)
}
