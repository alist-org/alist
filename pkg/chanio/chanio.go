package chanio

import (
	"io"
	"sync/atomic"
)

type ChanIO struct {
	cl  atomic.Bool
	c   chan []byte
	buf []byte
}

func New() *ChanIO {
	return &ChanIO{
		cl:  atomic.Bool{},
		c:   make(chan []byte),
		buf: make([]byte, 0),
	}
}

func (c *ChanIO) Read(p []byte) (int, error) {
	if c.cl.Load() {
		if len(c.buf) == 0 {
			return 0, io.EOF
		}
		n := copy(p, c.buf)
		if len(c.buf) > n {
			c.buf = c.buf[n:]
		} else {
			c.buf = make([]byte, 0)
		}
		return n, nil
	}
	for len(c.buf) < len(p) && !c.cl.Load() {
		c.buf = append(c.buf, <-c.c...)
	}
	n := copy(p, c.buf)
	if len(c.buf) > n {
		c.buf = c.buf[n:]
	} else {
		c.buf = make([]byte, 0)
	}
	return n, nil
}

func (c *ChanIO) Write(p []byte) (int, error) {
	if c.cl.Load() {
		return 0, io.ErrClosedPipe
	}
	c.c <- p
	return len(p), nil
}

func (c *ChanIO) Close() error {
	if c.cl.Load() {
		return io.ErrClosedPipe
	}
	c.cl.Store(true)
	close(c.c)
	return nil
}
