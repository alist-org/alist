package task

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
)

func TestTask_Manager(t *testing.T) {
	tm := NewTaskManager(3, func(id *uint64) {
		atomic.AddUint64(id, 1)
	})
	id := tm.Submit(WithCancelCtx(&Task[uint64]{
		Name: "test",
		Func: func(task *Task[uint64]) error {
			time.Sleep(time.Millisecond * 500)
			return nil
		},
	}))
	task, ok := tm.Get(id)
	if !ok {
		t.Fatal("task not found")
	}
	time.Sleep(time.Millisecond * 100)
	if task.state != RUNNING {
		t.Errorf("task status not running: %s", task.state)
	}
	time.Sleep(time.Second)
	if task.state != SUCCEEDED {
		t.Errorf("task status not finished: %s", task.state)
	}
}

func TestTask_Cancel(t *testing.T) {
	tm := NewTaskManager(3, func(id *uint64) {
		atomic.AddUint64(id, 1)
	})
	id := tm.Submit(WithCancelCtx(&Task[uint64]{
		Name: "test",
		Func: func(task *Task[uint64]) error {
			for {
				if utils.IsCanceled(task.Ctx) {
					return nil
				} else {
					t.Logf("task is running")
				}
			}
		},
	}))
	task, ok := tm.Get(id)
	if !ok {
		t.Fatal("task not found")
	}
	time.Sleep(time.Microsecond * 50)
	task.Cancel()
	time.Sleep(time.Millisecond)
	if task.state != CANCELED {
		t.Errorf("task status not canceled: %s", task.state)
	}
}

func TestTask_Retry(t *testing.T) {
	tm := NewTaskManager(3, func(id *uint64) {
		atomic.AddUint64(id, 1)
	})
	num := 0
	id := tm.Submit(WithCancelCtx(&Task[uint64]{
		Name: "test",
		Func: func(task *Task[uint64]) error {
			num++
			if num&1 == 1 {
				return errors.New("test error")
			}
			return nil
		},
	}))
	task, ok := tm.Get(id)
	if !ok {
		t.Fatal("task not found")
	}
	time.Sleep(time.Millisecond)
	if task.Error == nil {
		t.Error(task.state)
		t.Fatal("task error is nil, but expected error")
	} else {
		t.Logf("task error: %s", task.Error)
	}
	task.retry()
	time.Sleep(time.Millisecond)
	if task.Error != nil {
		t.Errorf("task error: %+v, but expected nil", task.Error)
	}
}
