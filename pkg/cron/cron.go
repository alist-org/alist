package cron

import "time"

type Cron struct {
	d  time.Duration
	ch chan struct{}
}

func NewCron(d time.Duration) *Cron {
	return &Cron{
		d:  d,
		ch: make(chan struct{}),
	}
}

func (c *Cron) Do(f func()) {
	go func() {
		ticker := time.NewTicker(c.d)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				f()
			case <-c.ch:
				return
			}
		}
	}()
}

func (c *Cron) Stop() {
	select {
	case _, _ = <-c.ch:
	default:
		c.ch <- struct{}{}
		close(c.ch)
	}
}
