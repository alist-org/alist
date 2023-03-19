package ftp

import (
	"io"
	"os"
	"sync"
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

// An FTP file reader that implements io.ReadSeekCloser for seeking.
type FTPFileReader struct {
	conn   *ftp.ServerConn
	resp   *ftp.Response
	offset int64
	mu     sync.Mutex
	path   string
}

func NewFTPFileReader(conn *ftp.ServerConn, path string) *FTPFileReader {
	return &FTPFileReader{
		conn: conn,
		path: path,
	}
}

func (r *FTPFileReader) Read(buf []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.resp == nil {
		r.resp, err = r.conn.RetrFrom(r.path, uint64(r.offset))
		if err != nil {
			return 0, err
		}
	}

	n, err = r.resp.Read(buf)
	r.offset += int64(n)
	return
}

func (r *FTPFileReader) Seek(offset int64, whence int) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	oldOffset := r.offset
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = oldOffset + offset
	case io.SeekEnd:
		size, err := r.conn.FileSize(r.path)
		if err != nil {
			return oldOffset, err
		}
		newOffset = offset + int64(size)
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
	r.offset = newOffset

	if r.resp != nil {
		// close the existing ftp data connection, otherwise the next read will be blocked
		_ = r.resp.Close() // we do not care about whether it returns an error
		r.resp = nil
	}
	return newOffset, nil
}

func (r *FTPFileReader) Close() error {
	if r.resp != nil {
		return r.resp.Close()
	}
	return nil
}
