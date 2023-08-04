package utils

import (
	"bytes"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"time"
)

// here is some syntaxic sugar inspired by the Tomas Senart's video,
// it allows me to inline the Reader interface
type readerFunc func(p []byte) (n int, err error)

func (rf readerFunc) Read(p []byte) (n int, err error) { return rf(p) }

// CopyWithCtx slightly modified function signature:
// - context has been added in order to propagate cancelation
// - I do not return the number of bytes written, has it is not useful in my use case
func CopyWithCtx(ctx context.Context, out io.Writer, in io.Reader, size int64, progress func(percentage int)) error {
	// Copy will call the Reader and Writer interface multiple time, in order
	// to copy by chunk (avoiding loading the whole file in memory).
	// I insert the ability to cancel before read time as it is the earliest
	// possible in the call process.
	var finish int64 = 0
	s := size / 100
	_, err := io.Copy(out, readerFunc(func(p []byte) (int, error) {
		// golang non-blocking channel: https://gobyexample.com/non-blocking-channel-operations
		select {
		// if context has been canceled
		case <-ctx.Done():
			// stop process and propagate "context canceled" error
			return 0, ctx.Err()
		default:
			// otherwise just run default io.Reader implementation
			n, err := in.Read(p)
			if s > 0 && (err == nil || err == io.EOF) {
				finish += int64(n)
				progress(int(finish / s))
			}
			return n, err
		}
	}))
	return err
}

type limitWriter struct {
	w     io.Writer
	count int64
	limit int64
}

func (l limitWriter) Write(p []byte) (n int, err error) {
	wn := int(l.limit - l.count)
	if wn > len(p) {
		wn = len(p)
	}
	if wn > 0 {
		if n, err = l.w.Write(p[:wn]); err != nil {
			return
		}
		if n < wn {
			err = io.ErrShortWrite
		}
	}
	if err == nil {
		n = len(p)
	}
	return
}

func LimitWriter(w io.Writer, size int64) io.Writer {
	return &limitWriter{w: w, limit: size}
}

type ReadCloser struct {
	io.Reader
	io.Closer
}

type CloseFunc func() error

func (c CloseFunc) Close() error {
	return c()
}

func NewReadCloser(reader io.Reader, close CloseFunc) io.ReadCloser {
	return ReadCloser{
		Reader: reader,
		Closer: close,
	}
}

func NewLimitReadCloser(reader io.Reader, close CloseFunc, limit int64) io.ReadCloser {
	return NewReadCloser(io.LimitReader(reader, limit), close)
}

type MultiReadable struct {
	originReader io.Reader
	reader       io.Reader
	cache        *bytes.Buffer
}

func NewMultiReadable(reader io.Reader) *MultiReadable {
	return &MultiReadable{
		originReader: reader,
		reader:       reader,
	}
}

func (mr *MultiReadable) Read(p []byte) (int, error) {
	n, err := mr.reader.Read(p)
	if _, ok := mr.reader.(io.Seeker); !ok && n > 0 {
		if mr.cache == nil {
			mr.cache = &bytes.Buffer{}
		}
		mr.cache.Write(p[:n])
	}
	return n, err
}

func (mr *MultiReadable) Reset() error {
	if seeker, ok := mr.reader.(io.Seeker); ok {
		_, err := seeker.Seek(0, io.SeekStart)
		return err
	}
	if mr.cache != nil && mr.cache.Len() > 0 {
		mr.reader = io.MultiReader(mr.cache, mr.reader)
		mr.cache = nil
	}
	return nil
}

func (mr *MultiReadable) Close() error {
	if closer, ok := mr.originReader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

type nopCloser struct {
	io.ReadSeeker
}

func (nopCloser) Close() error { return nil }

func ReadSeekerNopCloser(r io.ReadSeeker) io.ReadSeekCloser {
	return nopCloser{r}
}

func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		fmt.Println("This is attempt number", i)
		if i > 0 {
			log.Println("retrying after error:", err)
			time.Sleep(sleep)
			sleep *= 2
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

type Closers struct {
	closers []*io.Closer
}

func (c *Closers) Close() (err error) {
	for _, closer := range c.closers {
		if closer != nil {
			_ = (*closer).Close()
		}
	}
	return nil
}
func (c *Closers) Add(closer io.Closer) {
	if closer != nil {
		c.closers = append(c.closers, &closer)
	}
}
func NewClosers() *Closers {
	return &Closers{[]*io.Closer{}}
}
