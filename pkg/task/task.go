// Package task manage task, such as file upload, file copy between accounts, offline download, etc.
package task

import (
	"context"
	"github.com/pkg/errors"
)

var (
	PENDING   = "pending"
	RUNNING   = "running"
	FINISHED  = "finished"
	CANCELING = "canceling"
	CANCELED  = "canceled"
)

type Func func(task *Task) error

type Task struct {
	ID     uint64
	Name   string
	Status string
	Error  error
	Func   Func
	Ctx    context.Context
	cancel context.CancelFunc
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

func (t *Task) Run() {
	t.Status = RUNNING
	t.Error = t.Func(t)
	if errors.Is(t.Ctx.Err(), context.Canceled) {
		t.Status = CANCELED
	} else {
		t.Status = FINISHED
	}
}

func (t *Task) Retry() {
	t.Run()
}

func (t *Task) Cancel() {
	if t.cancel != nil {
		t.cancel()
	}
	// maybe can't cancel
	t.Status = CANCELING
}
