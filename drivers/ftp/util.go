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

// An FTP file reader that implements io.MFile for seeking.
type FTPFileReader struct {
	conn         *ftp.ServerConn
	resp         *ftp.Response
	offset       int64
	readAtOffset int64
	mu           sync.Mutex
	path         string
}

func NewFTPFileReader(conn *ftp.ServerConn, path string) *FTPFileReader {
	return &FTPFileReader{
		conn: conn,
		path: path,
	}
}

func (r *FTPFileReader) Read(buf []byte) (n int, err error) {
	n, err = r.ReadAt(buf, r.offset)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.offset += int64(n)
	return
}
func (r *FTPFileReader) ReadAt(buf []byte, off int64) (n int, err error) {
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
	return newOffset, nil
}

func (r *FTPFileReader) Close() error {
	if r.resp != nil {
		return r.resp.Close()
	}
	return nil
}
