package mega

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/t3rm1n4l/go-mega"
	"io"
	"sync"
	"time"
)

// do others that not defined in Driver interface
// openObject represents a download in progress
type openObject struct {
	ctx    context.Context
	mu     sync.Mutex
	d      *mega.Download
	id     int
	skip   int64
	chunk  []byte
	closed bool
}

// get the next chunk
func (oo *openObject) getChunk(ctx context.Context) (err error) {
	if oo.id >= oo.d.Chunks() {
		return io.EOF
	}
	var chunk []byte
	err = utils.Retry(3, time.Second, func() (err error) {
		chunk, err = oo.d.DownloadChunk(oo.id)
		return err
	})
	if err != nil {
		return err
	}
	oo.id++
	oo.chunk = chunk
	return nil
}

// Read reads up to len(p) bytes into p.
func (oo *openObject) Read(p []byte) (n int, err error) {
	oo.mu.Lock()
	defer oo.mu.Unlock()
	if oo.closed {
		return 0, fmt.Errorf("read on closed file")
	}
	// Skip data at the start if requested
	for oo.skip > 0 {
		_, size, err := oo.d.ChunkLocation(oo.id)
		if err != nil {
			return 0, err
		}
		if oo.skip < int64(size) {
			break
		}
		oo.id++
		oo.skip -= int64(size)
	}
	if len(oo.chunk) == 0 {
		err = oo.getChunk(oo.ctx)
		if err != nil {
			return 0, err
		}
		if oo.skip > 0 {
			oo.chunk = oo.chunk[oo.skip:]
			oo.skip = 0
		}
	}
	n = copy(p, oo.chunk)
	oo.chunk = oo.chunk[n:]
	return n, nil
}

// Close closed the file - MAC errors are reported here
func (oo *openObject) Close() (err error) {
	oo.mu.Lock()
	defer oo.mu.Unlock()
	if oo.closed {
		return nil
	}
	err = utils.Retry(3, 500*time.Millisecond, func() (err error) {
		return oo.d.Finish()
	})
	if err != nil {
		return fmt.Errorf("failed to finish download: %w", err)
	}
	oo.closed = true
	return nil
}
