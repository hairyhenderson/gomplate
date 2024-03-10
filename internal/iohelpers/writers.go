package iohelpers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hack-pad/hackpadfs"
)

type emptySkipper struct {
	open func() (io.Writer, error)

	// internal
	w   io.Writer
	buf *bytes.Buffer
	nw  bool
}

// NewEmptySkipper creates an io.WriteCloser that will only start writing once a
// non-whitespace byte has been encountered. The wrapped io.WriteCloser must be
// provided by the `open` func.
func NewEmptySkipper(open func() (io.Writer, error)) io.WriteCloser {
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
	if wc, ok := f.w.(io.WriteCloser); ok {
		return wc.Close()
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
func NopCloser(w io.Writer) io.WriteCloser {
	return &nopCloser{Writer: w}
}

type nopCloser struct {
	io.Writer
}

// Close - implements io.Closer
func (n *nopCloser) Close() error {
	return nil
}

var (
	_ io.WriteCloser = (*nopCloser)(nil)
	_ io.WriteCloser = (*emptySkipper)(nil)
	_ io.WriteCloser = (*sameSkipper)(nil)
)

type sameSkipper struct {
	open func() (io.WriteCloser, error)

	// internal
	r    *bufio.Reader
	w    io.WriteCloser
	buf  *bytes.Buffer
	diff bool
}

// SameSkipper creates an io.WriteCloser that will only start writing once a
// difference with the current output has been encountered. The wrapped
// io.WriteCloser must be provided by 'open'.
func SameSkipper(r io.Reader, open func() (io.WriteCloser, error)) io.WriteCloser {
	br := bufio.NewReader(r)
	return &sameSkipper{
		r:    br,
		w:    nil,
		buf:  &bytes.Buffer{},
		diff: false,
		open: open,
	}
}

// Write - writes to the buffer, until a difference with the output is found,
// then flushes and writes to the wrapped writer.
func (f *sameSkipper) Write(p []byte) (n int, err error) {
	if !f.diff {
		in := make([]byte, len(p))
		_, err := f.r.Read(in)
		if err != nil && err != io.EOF {
			return 0, fmt.Errorf("failed to read: %w", err)
		}
		if bytes.Equal(in, p) {
			return f.buf.Write(p)
		}

		f.diff = true
		err = f.flush()
		if err != nil {
			return 0, err
		}
	}
	return f.w.Write(p)
}

func (f *sameSkipper) flush() (err error) {
	if f.w == nil {
		f.w, err = f.open()
		if err != nil {
			return err
		}
		if f.w == nil {
			return fmt.Errorf("nil writer returned by open")
		}
	}
	// empty the buffer into the wrapped writer
	_, err = f.buf.WriteTo(f.w)
	return err
}

// Close - implements io.Closer
func (f *sameSkipper) Close() error {
	// Check to see if we missed anything in the reader
	if !f.diff {
		n, err := f.r.Peek(1)
		if len(n) > 0 || err != io.EOF {
			err = f.flush()
			if err != nil {
				return fmt.Errorf("failed to flush on close: %w", err)
			}
		}
	}

	if f.w != nil {
		return f.w.Close()
	}
	return nil
}

// LazyWriteCloser provides an interface to a WriteCloser that will open on the
// first access. The wrapped io.WriteCloser must be provided by 'open'.
func LazyWriteCloser(open func() (io.WriteCloser, error)) io.WriteCloser {
	return &lazyWriteCloser{
		opened: sync.Once{},
		open:   open,
	}
}

type lazyWriteCloser struct {
	w io.WriteCloser
	// caches the error that came from open(), if any
	openErr error
	open    func() (io.WriteCloser, error)
	opened  sync.Once
}

var _ io.WriteCloser = (*lazyWriteCloser)(nil)

func (l *lazyWriteCloser) openWriter() (r io.WriteCloser, err error) {
	l.opened.Do(func() {
		l.w, l.openErr = l.open()
	})
	return l.w, l.openErr
}

func (l *lazyWriteCloser) Close() error {
	w, err := l.openWriter()
	if err != nil {
		return err
	}
	return w.Close()
}

func (l *lazyWriteCloser) Write(p []byte) (n int, err error) {
	w, err := l.openWriter()
	if err != nil {
		return 0, err
	}
	return w.Write(p)
}

// WriteFile writes the given content to the file, truncating any existing file,
// and creating the directory structure leading up to it if necessary.
func WriteFile(fsys fs.FS, filename string, content []byte) error {
	err := assertPathInWD(filename)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filename, err)
	}

	fi, err := fs.Stat(fsys, filename)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to stat %s: %w", filename, err)
	}
	mode := NormalizeFileMode(0o644)
	if fi != nil {
		mode = fi.Mode()
	}

	err = hackpadfs.MkdirAll(fsys, filepath.Dir(filename), 0o755)
	if err != nil {
		return fmt.Errorf("failed to make dirs for %s: %w", filename, err)
	}

	err = hackpadfs.WriteFullFile(fsys, filename, content, mode)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}

	return nil
}

func assertPathInWD(filename string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	f, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	r, err := filepath.Rel(wd, f)
	if err != nil {
		return err
	}
	if strings.HasPrefix(r, "..") {
		return fmt.Errorf("path %s not contained by working directory %s (rel: %s)", filename, wd, r)
	}
	return nil
}
