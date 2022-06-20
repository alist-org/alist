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

type Func func(task *Task) error
type Callback func(task *Task)

type Task struct {
	ID       uint64
	Name     string
	Status   string
	Error    error
	Func     Func
	Ctx      context.Context
	progress int
	callback Callback
	cancel   context.CancelFunc
}

func newTask(name string, func_ Func, callbacks ...Callback) *Task {
	ctx, cancel := context.WithCancel(context.Background())
	t := &Task{
		Name:   name,
		Status: PENDING,
		Func:   func_,
		Ctx:    ctx,
		cancel: cancel,
	}
	if len(callbacks) > 0 {
		t.callback = callbacks[0]
	}
	return t
}

func (t *Task) SetStatus(status string) {
	t.Status = status
}

func (t *Task) SetProgress(percentage int) {
	t.progress = percentage
}

func (t *Task) run() {
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

func (t *Task) retry() {
	t.run()
}

func (t *Task) Cancel() {
	if t.Status == FINISHED || t.Status == CANCELED {
		return
	}
	if t.cancel != nil {
		t.cancel()
	}
	// maybe can't cancel
	t.Status = CANCELING
}
