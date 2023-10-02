package errgroup

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/avast/retry-go"
)

type token struct{}
type Group struct {
	cancel func(error)
	ctx    context.Context
	opts   []retry.Option

	success uint64

	wg  sync.WaitGroup
	sem chan token
}

func NewGroupWithContext(ctx context.Context, limit int, retryOpts ...retry.Option) (*Group, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return (&Group{cancel: cancel, ctx: ctx, opts: append(retryOpts, retry.Context(ctx))}).SetLimit(limit), ctx
}

func (g *Group) done() {
	if g.sem != nil {
		<-g.sem
	}
	g.wg.Done()
	atomic.AddUint64(&g.success, 1)
}

func (g *Group) Wait() error {
	g.wg.Wait()
	return context.Cause(g.ctx)
}

func (g *Group) Go(f func(ctx context.Context) error) {
	if g.sem != nil {
		g.sem <- token{}
	}

	g.wg.Add(1)
	go func() {
		defer g.done()
		if err := retry.Do(func() error { return f(g.ctx) }, g.opts...); err != nil {
			g.cancel(err)
		}
	}()
}

func (g *Group) TryGo(f func(ctx context.Context) error) bool {
	if g.sem != nil {
		select {
		case g.sem <- token{}:
		default:
			return false
		}
	}

	g.wg.Add(1)
	go func() {
		defer g.done()
		if err := retry.Do(func() error { return f(g.ctx) }, g.opts...); err != nil {
			g.cancel(err)
		}
	}()
	return true
}

func (g *Group) SetLimit(n int) *Group {
	if len(g.sem) != 0 {
		panic(fmt.Errorf("errgroup: modify limit while %v goroutines in the group are still active", len(g.sem)))
	}
	if n > 0 {
		g.sem = make(chan token, n)
	} else {
		g.sem = nil
	}
	return g
}

func (g *Group) Success() uint64 {
	return atomic.LoadUint64(&g.success)
}

func (g *Group) Err() error {
	return context.Cause(g.ctx)
}
