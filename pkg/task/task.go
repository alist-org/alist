// Package task manage task, such as file upload, file copy between accounts, offline download, etc.
package task

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	PENDING   = "pending"
	RUNNING   = "running"
	FINISHED  = "finished"
	CANCELING = "canceling"
	CANCELED  = "canceled"
	ERRORED   = "errored"
)

type Func[K comparable, V any] func(task *Task[K, V]) error
type Callback[K comparable, V any] func(task *Task[K, V])

type Task[K comparable, V any] struct {
	ID     K
	Name   string
	Status string
	Error  error

	Data V

	Func     Func[K, V]
	callback Callback[K, V]

	Ctx      context.Context
	progress int
	cancel   context.CancelFunc
}

func (t *Task[K, V]) SetStatus(status string) {
	t.Status = status
}

func (t *Task[K, V]) SetProgress(percentage int) {
	t.progress = percentage
}

func (t *Task[K, V]) run() {
	t.Status = RUNNING
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("error [%+v] while run task [%s]", err, t.Name)
			t.Error = errors.Errorf("panic: %+v", err)
			t.Status = ERRORED
		}
	}()
	t.Error = t.Func(t)
	if errors.Is(t.Ctx.Err(), context.Canceled) {
		t.Status = CANCELED
	} else if t.Error != nil {
		t.Status = ERRORED
	} else {
		t.Status = FINISHED
		if t.callback != nil {
			t.callback(t)
		}
	}
}

func (t *Task[K, V]) retry() {
	t.run()
}

func (t *Task[K, V]) Cancel() {
	if t.Status == FINISHED || t.Status == CANCELED {
		return
	}
	if t.cancel != nil {
		t.cancel()
	}
	// maybe can't cancel
	t.Status = CANCELING
}

func WithCancelCtx[K comparable, V any](task *Task[K, V]) *Task[K, V] {
	ctx, cancel := context.WithCancel(context.Background())
	task.Ctx = ctx
	task.cancel = cancel
	task.Status = PENDING
	return task
}
