package writers

import (
	"bytes"
	"errors"
	"io"
)

type emptySkipper struct {
	open func() (io.WriteCloser, error)

	// internal
	w   io.WriteCloser
	buf *bytes.Buffer
	nw  bool
}

// NewEmptySkipper creates an io.WriteCloser that will only start writing once a
// non-whitespace byte has been encountered. The wrapped io.WriteCloser must be
// provided by the `open` func.
func NewEmptySkipper(open func() (io.WriteCloser, error)) io.WriteCloser {
	return &emptySkipper{
		w:    nil,
		buf:  &bytes.Buffer{},
		nw:   false,
		open: open,
	}
}

func (f *emptySkipper) Write(p []byte) (n int, err error) {
	if !f.nw {
		if allWhitespace(p) {
			// buffer the whitespace
			return f.buf.Write(p)
		}

		// first time around, so open the writer
		f.nw = true
		f.w, err = f.open()
		if err != nil {
			return 0, err
		}
		if f.w == nil {
			return 0, errors.New("nil writer returned by open")
		}
		// empty the buffer into the wrapped writer
		_, err = f.buf.WriteTo(f.w)
		if err != nil {
			return 0, err
		}
	}

	return f.w.Write(p)
}

// Close - implements io.Closer
func (f *emptySkipper) Close() error {
	if f.w != nil {
		return f.w.Close()
	}
	return nil
}

func allWhitespace(p []byte) bool {
	for _, b := range p {
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\v' {
			continue
		}
		return false
	}
	return true
}

// NopCloser returns a WriteCloser with a no-op Close method wrapping
// the provided io.Writer.
type NopCloser struct {
	io.Writer
}

// Close - implements io.Closer
func (n *NopCloser) Close() error {
	return nil
}

var (
	_ io.WriteCloser = (*NopCloser)(nil)
	_ io.WriteCloser = (*emptySkipper)(nil)
)
