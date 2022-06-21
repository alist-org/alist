package task

import (
	"github.com/alist-org/alist/v3/pkg/generic_sync"
	log "github.com/sirupsen/logrus"
)

type Manager[K comparable, V any] struct {
	workerC  chan struct{}
	curID    K
	updateID func(*K)
	tasks    generic_sync.MapOf[K, *Task[K, V]]
}

func (tm *Manager[K, V]) Submit(task *Task[K, V]) K {
	if tm.updateID != nil {
		task.ID = tm.curID
		tm.updateID(&task.ID)
	}
	tm.tasks.Store(task.ID, task)
	tm.do(task)
	return task.ID
}

func (tm *Manager[K, V]) do(task *Task[K, V]) {
	go func() {
		log.Debugf("task [%s] waiting for worker", task.Name)
		select {
		case <-tm.workerC:
			log.Debugf("task [%s] starting", task.Name)
			task.run()
			log.Debugf("task [%s] ended", task.Name)
		}
		// return worker
		tm.workerC <- struct{}{}
	}()
}

func (tm *Manager[K, V]) GetAll() []*Task[K, V] {
	return tm.tasks.Values()
}

func (tm *Manager[K, V]) Get(tid K) (*Task[K, V], bool) {
	return tm.tasks.Load(tid)
}

func (tm *Manager[K, V]) MustGet(tid K) *Task[K, V] {
	task, _ := tm.Get(tid)
	return task
}

func (tm *Manager[K, V]) Retry(tid K) error {
	t, ok := tm.Get(tid)
	if !ok {
		return ErrTaskNotFound
	}
	tm.do(t)
	return nil
}

func (tm *Manager[K, V]) Cancel(tid K) error {
	t, ok := tm.Get(tid)
	if !ok {
		return ErrTaskNotFound
	}
	t.Cancel()
	return nil
}

func (tm *Manager[K, V]) Remove(tid K) {
	tm.tasks.Delete(tid)
}

// RemoveAll removes all tasks from the manager, this maybe shouldn't be used
// because the task maybe still running.
func (tm *Manager[K, V]) RemoveAll() {
	tm.tasks.Clear()
}

func (tm *Manager[K, V]) RemoveFinished() {
	tasks := tm.GetAll()
	for _, task := range tasks {
		if task.Status == FINISHED {
			tm.Remove(task.ID)
		}
	}
}

func (tm *Manager[K, V]) RemoveError() {
	tasks := tm.GetAll()
	for _, task := range tasks {
		if task.Error != nil {
			tm.Remove(task.ID)
		}
	}
}

func NewTaskManager[K comparable, V any](maxWorker int, updateID ...func(*K)) *Manager[K, V] {
	tm := &Manager[K, V]{
		tasks:   generic_sync.MapOf[K, *Task[K, V]]{},
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
