// Package task manage task, such as file upload, file copy between storages, offline download, etc.
package task

import (
	"context"
	"runtime"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	PENDING   = "pending"
	RUNNING   = "running"
	SUCCEEDED = "succeeded"
	CANCELING = "canceling"
	CANCELED  = "canceled"
	ERRORED   = "errored"
)

type Func[K comparable] func(task *Task[K]) error
type Callback[K comparable] func(task *Task[K])

type Task[K comparable] struct {
	ID       K
	Name     string
	state    string // pending, running, finished, canceling, canceled, errored
	status   string
	progress float64

	Error error

	Func     Func[K]
	callback Callback[K]

	Ctx    context.Context
	cancel context.CancelFunc
}

func (t *Task[K]) SetStatus(status string) {
	t.status = status
}

func (t *Task[K]) SetProgress(percentage float64) {
	t.progress = percentage
}

func (t Task[K]) GetProgress() float64 {
	return t.progress
}

func (t Task[K]) GetState() string {
	return t.state
}

func (t Task[K]) GetStatus() string {
	return t.status
}

func (t Task[K]) GetErrMsg() string {
	if t.Error == nil {
		return ""
	}
	return t.Error.Error()
}

func getCurrentGoroutineStack() string {
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func (t *Task[K]) run() {
	t.state = RUNNING
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("error [%s] while run task [%s],stack trace:\n%s", err, t.Name, getCurrentGoroutineStack())
			t.Error = errors.Errorf("panic: %+v", err)
			t.state = ERRORED
		}
	}()
	t.Error = t.Func(t)
	if t.Error != nil {
		log.Errorf("error [%+v] while run task [%s]", t.Error, t.Name)
	}
	if errors.Is(t.Ctx.Err(), context.Canceled) {
		t.state = CANCELED
	} else if t.Error != nil {
		t.state = ERRORED
	} else {
		t.state = SUCCEEDED
		t.SetProgress(100)
		if t.callback != nil {
			t.callback(t)
		}
	}
}

func (t *Task[K]) retry() {
	t.run()
}

func (t *Task[K]) Done() bool {
	return t.state == SUCCEEDED || t.state == CANCELED || t.state == ERRORED
}

func (t *Task[K]) Cancel() {
	if t.state == SUCCEEDED || t.state == CANCELED {
		return
	}
	if t.cancel != nil {
		t.cancel()
	}
	// maybe can't cancel
	t.state = CANCELING
}

func WithCancelCtx[K comparable](task *Task[K]) *Task[K] {
	ctx, cancel := context.WithCancel(context.Background())
	task.Ctx = ctx
	task.cancel = cancel
	task.state = PENDING
	return task
}
