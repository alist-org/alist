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

type Func[K comparable] func(task *Task[K]) error
type Callback[K comparable] func(task *Task[K])

type Task[K comparable] struct {
	ID     K
	Name   string
	Status string
	Error  error

	Func     Func[K]
	callback Callback[K]

	Ctx      context.Context
	progress int
	cancel   context.CancelFunc
}

func (t *Task[K]) SetStatus(status string) {
	t.Status = status
}

func (t *Task[K]) SetProgress(percentage int) {
	t.progress = percentage
}

func (t *Task[K]) run() {
	t.Status = RUNNING
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("error [%+v] while run task [%s]", err, t.Name)
			t.Error = errors.Errorf("panic: %+v", err)
			t.Status = ERRORED
		}
	}()
	t.Error = t.Func(t)
	if t.Error != nil {
		log.Errorf("error [%+v] while run task [%s]", t.Error, t.Name)
	}
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

func (t *Task[K]) retry() {
	t.run()
}

func (t *Task[K]) Cancel() {
	if t.Status == FINISHED || t.Status == CANCELED {
		return
	}
	if t.cancel != nil {
		t.cancel()
	}
	// maybe can't cancel
	t.Status = CANCELING
}

func WithCancelCtx[K comparable](task *Task[K]) *Task[K] {
	ctx, cancel := context.WithCancel(context.Background())
	task.Ctx = ctx
	task.cancel = cancel
	task.Status = PENDING
	return task
}
