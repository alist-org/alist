package cron

import (
	"testing"
	"time"
)

func TestCron(t *testing.T) {
	c := NewCron(time.Second)
	c.Do(func() {
		t.Logf("cron log")
	})
	time.Sleep(time.Second * 5)
	c.Stop()
}
