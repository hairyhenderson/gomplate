package memfs

import (
	"sync"
)

// MemFile represents a file backed by a Buffer which is secured from concurrent access.
type MemFile struct {
	Buffer
	mutex *sync.RWMutex
	name  string
}

// NewMemFile creates a Buffer which byte slice is safe from concurrent access,
// the file itself is not thread-safe.
//
// This means multiple files can work safely on the same byte slice,
// but multiple go routines working on the same file may corrupt the internal pointer structure.
func NewMemFile(name string, rwMutex *sync.RWMutex, buf *[]byte) *MemFile {
	return &MemFile{
		Buffer: NewBuffer(buf),
		mutex:  rwMutex,
		name:   name,
	}
}

// Name of the file
func (b MemFile) Name() string {
	return b.name
}

// Sync has no effect
func (b MemFile) Sync() error {
	return nil
}

// Truncate changes the size of the file
func (b MemFile) Truncate(size int64) (err error) {
	b.mutex.Lock()
	err = b.Buffer.Truncate(size)
	b.mutex.Unlock()
	return
}

// Read reads len(p) byte from the underlying buffer starting at the current offset.
// It returns the number of bytes read and an error if any.
// Returns io.EOF error if pointer is at the end of the Buffer.
// See Buf.Read()
func (b *MemFile) Read(p []byte) (n int, err error) {
	b.mutex.RLock()
	n, err = b.Buffer.Read(p)
	b.mutex.RUnlock()
	return
}

// ReadAt reads len(b) bytes from the Buffer starting at byte offset off.
// It returns the number of bytes read and the error, if any.
// ReadAt always returns a non-nil error when n < len(b).
// At end of file, that error is io.EOF.
// See Buf.ReadAt()
func (b *MemFile) ReadAt(p []byte, off int64) (n int, err error) {
	b.mutex.RLock()
	n, err = b.Buffer.ReadAt(p, off)
	b.mutex.RUnlock()
	return
}

// Write writes len(p) byte to the Buffer.
// It returns the number of bytes written and an error if any.
// Write returns non-nil error when n!=len(p).
func (b *MemFile) Write(p []byte) (n int, err error) {
	b.mutex.Lock()
	n, err = b.Buffer.Write(p)
	b.mutex.Unlock()
	return
}

// Seek sets the offset for the next Read or Write on the buffer to offset,
// interpreted according to whence:
// 	0 (os.SEEK_SET) means relative to the origin of the file
// 	1 (os.SEEK_CUR) means relative to the current offset
// 	2 (os.SEEK_END) means relative to the end of the file
// It returns the new offset and an error, if any.
func (b *MemFile) Seek(offset int64, whence int) (n int64, err error) {
	b.mutex.RLock()
	n, err = b.Buffer.Seek(offset, whence)
	b.mutex.RUnlock()
	return
}
