package task

import (
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	"testing"
	"time"
)

func TestTask_Manager(t *testing.T) {
	tm := NewTaskManager()
	id := tm.Add("test", func(task *Task) error {
		time.Sleep(time.Millisecond * 500)
		return nil
	})
	task, ok := tm.Get(id)
	if !ok {
		t.Fatal("task not found")
	}
	time.Sleep(time.Millisecond * 100)
	if task.Status != RUNNING {
		t.Errorf("task status not running: %s", task.Status)
	}
	time.Sleep(time.Second)
	if task.Status != FINISHED {
		t.Errorf("task status not finished: %s", task.Status)
	}
}

func TestTask_Cancel(t *testing.T) {
	tm := NewTaskManager()
	id := tm.Add("test", func(task *Task) error {
		for {
			if utils.IsCanceled(task.Ctx) {
				return nil
			} else {
				t.Logf("task is running")
			}
		}
	})
	task, ok := tm.Get(id)
	if !ok {
		t.Fatal("task not found")
	}
	time.Sleep(time.Microsecond * 50)
	task.Cancel()
	time.Sleep(time.Millisecond)
	if task.Status != CANCELED {
		t.Errorf("task status not canceled: %s", task.Status)
	}
}

func TestTask_Retry(t *testing.T) {
	tm := NewTaskManager()
	num := 0
	id := tm.Add("test", func(task *Task) error {
		num++
		if num&1 == 1 {
			return errors.New("test error")
		}
		return nil
	})
	task, ok := tm.Get(id)
	if !ok {
		t.Fatal("task not found")
	}
	time.Sleep(time.Millisecond)
	if task.Error == nil {
		t.Error(task.Status)
		t.Fatal("task error is nil, but expected error")
	} else {
		t.Logf("task error: %s", task.Error)
	}
	task.Retry()
	time.Sleep(time.Millisecond)
	if task.Error != nil {
		t.Errorf("task error: %+v, but expected nil", task.Error)
	}
}
