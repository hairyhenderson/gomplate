package iohelpers

import (
	"io"
	"sync"
)

// LazyReadCloser provides an interface to a ReadCloser that will open on the
// first access. The wrapped io.ReadCloser must be provided by 'open'.
func LazyReadCloser(open func() (io.ReadCloser, error)) io.ReadCloser {
	return &lazyReadCloser{
		opened: sync.Once{},
		open:   open,
	}
}

type lazyReadCloser struct {
	r io.ReadCloser
	// caches the error that came from open(), if any
	openErr error
	open    func() (io.ReadCloser, error)
	opened  sync.Once
}

var _ io.ReadCloser = (*lazyReadCloser)(nil)

func (l *lazyReadCloser) openReader() (r io.ReadCloser, err error) {
	l.opened.Do(func() {
		l.r, l.openErr = l.open()
	})
	return l.r, l.openErr
}

func (l *lazyReadCloser) Close() error {
	r, err := l.openReader()
	if err != nil {
		return err
	}
	return r.Close()
}

func (l *lazyReadCloser) Read(p []byte) (n int, err error) {
	r, err := l.openReader()
	if err != nil {
		return 0, err
	}
	return r.Read(p)
}
