package base

import "io"

type Closers struct {
	closers []io.Closer
}

func (c *Closers) Close() (err error) {
	for _, closer := range c.closers {
		if closer != nil {
			_ = closer.Close()
		}
	}
	return nil
}
func (c *Closers) Add(closer io.Closer) {
	c.closers = append(c.closers, closer)
}
