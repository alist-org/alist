package local

import "context"

type TokenBucket interface {
	Take() <-chan struct{}
	Put()
	Do(context.Context, func() error) error
}

// StaticTokenBucket is a bucket with a fixed number of tokens,
// where the retrieval and return of tokens are manually controlled.
// In the initial state, the bucket is full.
type StaticTokenBucket struct {
	bucket chan struct{}
}

func NewStaticTokenBucket(size int) StaticTokenBucket {
	bucket := make(chan struct{}, size)
	for range size {
		bucket <- struct{}{}
	}
	return StaticTokenBucket{bucket: bucket}
}

func (b StaticTokenBucket) Take() <-chan struct{} {
	return b.bucket
}

func (b StaticTokenBucket) Put() {
	b.bucket <- struct{}{}
}

func (b StaticTokenBucket) Do(ctx context.Context, f func() error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.bucket:
		defer b.Put()
	}
	return f()
}

// NopTokenBucket all function calls to this bucket will success immediately
type NopTokenBucket struct {
	nop chan struct{}
}

func NewNopTokenBucket() NopTokenBucket {
	nop := make(chan struct{})
	close(nop)
	return NopTokenBucket{nop}
}

func (b NopTokenBucket) Take() <-chan struct{} {
	return b.nop
}

func (b NopTokenBucket) Put() {}

func (b NopTokenBucket) Do(_ context.Context, f func() error) error { return f() }
