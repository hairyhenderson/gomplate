package writers

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllWhitespace(t *testing.T) {
	testdata := []struct {
		in       []byte
		expected bool
	}{
		{[]byte(" "), true},
		{[]byte("foo"), false},
		{[]byte("   \t\n\n\v\r\n"), true},
		{[]byte("   foo   "), false},
	}

	for _, d := range testdata {
		assert.Equal(t, d.expected, allWhitespace(d.in))
	}
}

func TestEmptySkipper(t *testing.T) {
	testdata := []struct {
		in    []byte
		empty bool
	}{
		{[]byte(" "), true},
		{[]byte("foo"), false},
		{[]byte("   \t\n\n\v\r\n"), true},
		{[]byte("   foo   "), false},
	}

	for _, d := range testdata {
		w := &bufferCloser{&bytes.Buffer{}}
		opened := false
		f, ok := NewEmptySkipper(func() (io.WriteCloser, error) {
			opened = true
			return w, nil
		}).(*emptySkipper)

		assert.True(t, ok)
		n, err := f.Write(d.in)
		assert.NoError(t, err)
		assert.Equal(t, len(d.in), n)
		if d.empty {
			assert.Nil(t, f.w)
			assert.False(t, opened)
		} else {
			assert.NotNil(t, f.w)
			assert.True(t, opened)
			assert.EqualValues(t, d.in, w.Bytes())
		}
	}
}

type bufferCloser struct {
	*bytes.Buffer
}

func (b *bufferCloser) Close() error {
	return nil
}
