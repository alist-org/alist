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

type Task struct {
	ID       uint64
	Name     string
	Status   string
	Error    error
	Func     Func
	Progress int
	Ctx      context.Context
	cancel   context.CancelFunc
}

func newTask(name string, func_ Func) *Task {
	ctx, cancel := context.WithCancel(context.Background())
	return &Task{
		Name:   name,
		Status: PENDING,
		Func:   func_,
		Ctx:    ctx,
		cancel: cancel,
	}
}

func (t *Task) SetStatus(status string) {
	t.Status = status
}

func (t *Task) SetProgress(percentage int) {
	t.Progress = percentage
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
	}
}

func (t *Task) retry() {
	t.run()
}

func (t *Task) Cancel() {
	if t.cancel != nil {
		t.cancel()
	}
	// maybe can't cancel
	t.Status = CANCELING
}
