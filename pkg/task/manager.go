package task

import (
	"github.com/alist-org/alist/v3/pkg/generic_sync"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Manager[K comparable] struct {
	curID    K
	workerC  chan struct{}
	updateID func(*K)
	tasks    generic_sync.MapOf[K, *Task[K]]
}

func (tm *Manager[K]) Submit(task *Task[K]) K {
	if tm.updateID != nil {
		tm.updateID(&tm.curID)
		task.ID = tm.curID
	}
	tm.tasks.Store(task.ID, task)
	tm.do(task)
	return task.ID
}

func (tm *Manager[K]) do(task *Task[K]) {
	go func() {
		log.Debugf("task [%s] waiting for worker", task.Name)
		select {
		case <-tm.workerC:
			log.Debugf("task [%s] starting", task.Name)
			task.run()
			log.Debugf("task [%s] ended", task.Name)
		case <-task.Ctx.Done():
			log.Debugf("task [%s] canceled", task.Name)
			return
		}
		// return worker
		tm.workerC <- struct{}{}
	}()
}

func (tm *Manager[K]) GetAll() []*Task[K] {
	return tm.tasks.Values()
}

func (tm *Manager[K]) Get(tid K) (*Task[K], bool) {
	return tm.tasks.Load(tid)
}

func (tm *Manager[K]) MustGet(tid K) *Task[K] {
	task, _ := tm.Get(tid)
	return task
}

func (tm *Manager[K]) Retry(tid K) error {
	t, ok := tm.Get(tid)
	if !ok {
		return errors.WithStack(ErrTaskNotFound)
	}
	tm.do(t)
	return nil
}

func (tm *Manager[K]) Cancel(tid K) error {
	t, ok := tm.Get(tid)
	if !ok {
		return errors.WithStack(ErrTaskNotFound)
	}
	t.Cancel()
	return nil
}

func (tm *Manager[K]) Remove(tid K) error {
	t, ok := tm.Get(tid)
	if !ok {
		return errors.WithStack(ErrTaskNotFound)
	}
	if !t.Done() {
		return errors.WithStack(ErrTaskRunning)
	}
	tm.tasks.Delete(tid)
	return nil
}

// RemoveAll removes all tasks from the manager, this maybe shouldn't be used
// because the task maybe still running.
func (tm *Manager[K]) RemoveAll() {
	tm.tasks.Clear()
}

func (tm *Manager[K]) RemoveByStates(states ...string) {
	tasks := tm.GetAll()
	for _, task := range tasks {
		if utils.SliceContains(states, task.GetState()) {
			_ = tm.Remove(task.ID)
		}
	}
}

func (tm *Manager[K]) GetByStates(states ...string) []*Task[K] {
	var tasks []*Task[K]
	tm.tasks.Range(func(key K, value *Task[K]) bool {
		if utils.SliceContains(states, value.GetState()) {
			tasks = append(tasks, value)
		}
		return true
	})
	return tasks
}

func (tm *Manager[K]) ListUndone() []*Task[K] {
	return tm.GetByStates(PENDING, RUNNING, CANCELING)
}

func (tm *Manager[K]) ListDone() []*Task[K] {
	return tm.GetByStates(SUCCEEDED, CANCELED, ERRORED)
}

func (tm *Manager[K]) ClearDone() {
	tm.RemoveByStates(SUCCEEDED, CANCELED, ERRORED)
}

func (tm *Manager[K]) ClearSucceeded() {
	tm.RemoveByStates(SUCCEEDED)
}

func (tm *Manager[K]) RawTasks() *generic_sync.MapOf[K, *Task[K]] {
	return &tm.tasks
}

func NewTaskManager[K comparable](maxWorker int, updateID ...func(*K)) *Manager[K] {
	tm := &Manager[K]{
		tasks:   generic_sync.MapOf[K, *Task[K]]{},
		workerC: make(chan struct{}, maxWorker),
	}
	for i := 0; i < maxWorker; i++ {
		tm.workerC <- struct{}{}
	}
	if len(updateID) > 0 {
		tm.updateID = updateID[0]
	}
	return tm
}
