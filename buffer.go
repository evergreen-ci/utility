package utility

import (
	"bytes"
	"sync"
)

// MakeSafeBuffer returns a thread-safe Read/Write closer that wraps an existing
// bytes buffer.
func MakeSafeBuffer(b bytes.Buffer) *SafeBuffer { return &SafeBuffer{buffer: b} }

// SafeBuffer provides a thread-safe in-memory buffer.
type SafeBuffer struct {
	buffer bytes.Buffer
	sync.RWMutex
}

// Read performs a thread-safe in-memory read.
func (b *SafeBuffer) Read(p []byte) (n int, err error) {
	b.RLock()
	defer b.RUnlock()
	return b.buffer.Read(p)
}

// Write performs a thread-safe in-memory write.
func (b *SafeBuffer) Write(p []byte) (n int, err error) {
	b.Lock()
	defer b.Unlock()
	return b.buffer.Write(p)
}

// String returns the in-memory buffer contents as a string in a thread-safe
// manner.
func (b *SafeBuffer) String() string {
	b.RLock()
	defer b.RUnlock()
	return b.buffer.String()
}

// Close is a no-op to satisfy the closer interface.
func (b *SafeBuffer) Close() error { return nil }
