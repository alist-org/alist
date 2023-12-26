package ftp

import (
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jlaffaye/ftp"
)

// do others that not defined in Driver interface

func (d *FTP) login() error {
	if d.conn != nil {
		_, err := d.conn.CurrentDir()
		if err == nil {
			return nil
		}
	}
	conn, err := ftp.Dial(d.Address, ftp.DialWithShutTimeout(10*time.Second))
	if err != nil {
		return err
	}
	err = conn.Login(d.Username, d.Password)
	if err != nil {
		return err
	}
	d.conn = conn
	return nil
}

// FileReader An FTP file reader that implements io.MFile for seeking.
type FileReader struct {
	conn         *ftp.ServerConn
	resp         *ftp.Response
	offset       atomic.Int64
	readAtOffset int64
	mu           sync.Mutex
	path         string
	size         int64
}

func NewFileReader(conn *ftp.ServerConn, path string, size int64) *FileReader {
	return &FileReader{
		conn: conn,
		path: path,
		size: size,
	}
}

func (r *FileReader) Read(buf []byte) (n int, err error) {
	n, err = r.ReadAt(buf, r.offset.Load())
	r.offset.Add(int64(n))
	return
}

func (r *FileReader) ReadAt(buf []byte, off int64) (n int, err error) {
	if off < 0 {
		return -1, os.ErrInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if off != r.readAtOffset {
		//have to restart the connection, to correct offset
		_ = r.resp.Close()
		r.resp = nil
	}

	if r.resp == nil {
		r.resp, err = r.conn.RetrFrom(r.path, uint64(off))
		r.readAtOffset = off
		if err != nil {
			return 0, err
		}
	}

	n, err = r.resp.Read(buf)
	r.readAtOffset += int64(n)
	return
}

func (r *FileReader) Seek(offset int64, whence int) (int64, error) {
	oldOffset := r.offset.Load()
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = oldOffset + offset
	case io.SeekEnd:
		return r.size, nil
	default:
		return -1, os.ErrInvalid
	}

	if newOffset < 0 {
		// offset out of range
		return oldOffset, os.ErrInvalid
	}
	if newOffset == oldOffset {
		// offset not changed, so return directly
		return oldOffset, nil
	}
	r.offset.Store(newOffset)
	return newOffset, nil
}

func (r *FileReader) Close() error {
	if r.resp != nil {
		return r.resp.Close()
	}
	return nil
}
