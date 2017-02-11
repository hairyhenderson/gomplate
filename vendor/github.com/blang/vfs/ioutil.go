package vfs

import (
	"bytes"
	"io"
	"os"
)

// WriteFile writes data to a file named by filename on the given Filesystem. If
// the file does not exist, WriteFile creates it with permissions perm;
// otherwise WriteFile truncates it before writing.
//
// This is a port of the stdlib ioutil.WriteFile function.
func WriteFile(fs Filesystem, filename string, data []byte, perm os.FileMode) error {
	f, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

// ReadFile reads the file named by filename and returns the contents. A
// successful call returns err == nil, not err == EOF. Because ReadFile reads
// the whole file, it does not treat an EOF from Read as an error to be
// reported.
//
// This is a port of the stdlib ioutil.ReadFile function.
func ReadFile(fs Filesystem, filename string) ([]byte, error) {
	f, err := fs.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// It's a good but not certain bet that FileInfo will tell us exactly how
	// much to read, so let's try it but be prepared for the answer to be wrong.
	var n int64
	if fi, err := fs.Stat(filename); err == nil {
		if size := fi.Size(); size < 1e9 {
			n = size
		}
	}

	// As initial capacity for readAll, use n + a little extra in case Size is
	// zero, and to avoid another allocation after Read has filled the buffer.
	// The readAll call will read into its allocated internal buffer cheaply. If
	// the size was wrong, we'll either waste some space off the end or
	// reallocate as needed, but in the overwhelmingly common case we'll get it
	// just right.
	return readAll(f, n+bytes.MinRead)
}

// readAll reads from r until an error or EOF and returns the data it read from
// the internal buffer allocated with a specified capacity.
//
// This is a paste of the stdlib ioutil.readAll function.
func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, capacity))

	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()

	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}
