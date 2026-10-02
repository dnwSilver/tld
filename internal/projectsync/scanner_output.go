package projectsync

import (
	"bytes"
	"errors"
	"sync"
)

const maxScannerOutput = 64 << 20

var errScannerOutputLimit = errors.New("scanner output exceeds 64 MiB")

// scannerOutput bounds memory used by stdout and stderr from external tools.
// Both pipes may write concurrently, so the buffer is synchronized.
type scannerOutput struct {
	mu  sync.Mutex
	buf bytes.Buffer
	max int
}

func (b *scannerOutput) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.max <= 0 {
		b.max = maxScannerOutput
	}
	if len(data) > b.max-b.buf.Len() {
		return 0, errScannerOutputLimit
	}
	return b.buf.Write(data)
}

func (b *scannerOutput) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return bytes.Clone(b.buf.Bytes())
}
