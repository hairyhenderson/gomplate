package writers

import (
	"bytes"
	"fmt"
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
		err = f.Close()
		assert.NoError(t, err)
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

func TestSameSkipper(t *testing.T) {
	testdata := []struct {
		in   []byte
		out  []byte
		same bool
	}{
		{[]byte(" "), []byte(" "), true},
		{[]byte("foo"), []byte("foo"), true},
		{[]byte("foo"), nil, false},
		{[]byte("foo"), []byte("bar"), false},
		{[]byte("foobar"), []byte("foo"), false},
		{[]byte("foo"), []byte("foobar"), false},
	}

	for _, d := range testdata {
		t.Run(fmt.Sprintf("in:%q/out:%q/same:%v", d.in, d.out, d.same), func(t *testing.T) {
			r := bytes.NewBuffer(d.out)
			w := &bufferCloser{&bytes.Buffer{}}
			opened := false
			f, ok := SameSkipper(r, func() (io.WriteCloser, error) {
				opened = true
				return w, nil
			}).(*sameSkipper)
			assert.True(t, ok)

			n, err := f.Write(d.in)
			assert.NoError(t, err)
			assert.Equal(t, len(d.in), n)
			err = f.Close()
			assert.NoError(t, err)
			if d.same {
				assert.Nil(t, f.w)
				assert.False(t, opened)
				assert.Empty(t, w.Bytes())
			} else {
				assert.NotNil(t, f.w)
				assert.True(t, opened)
				assert.EqualValues(t, d.in, w.Bytes())
			}
		})
	}
}
