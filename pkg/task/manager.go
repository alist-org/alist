package task

import (
	log "github.com/sirupsen/logrus"
	"sync/atomic"

	"github.com/alist-org/alist/v3/pkg/generic_sync"
)

type Manager struct {
	workerC chan struct{}
	curID   uint64
	tasks   generic_sync.MapOf[uint64, *Task]
}

func (tm *Manager) Submit(name string, f Func) uint64 {
	task := newTask(name, f)
	tm.addTask(task)
	tm.do(task.ID)
	return task.ID
}

func (tm *Manager) do(tid uint64) {
	task := tm.MustGet(tid)
	go func() {
		log.Debugf("task [%s] waiting for worker", task.Name)
		select {
		case <-tm.workerC:
			log.Debugf("task [%s] starting", task.Name)
			task.run()
			log.Debugf("task [%s] ended", task.Name)
		}
		tm.workerC <- struct{}{}
	}()
}

func (tm *Manager) addTask(task *Task) {
	task.ID = tm.curID
	atomic.AddUint64(&tm.curID, 1)
	tm.tasks.Store(task.ID, task)
}

func (tm *Manager) GetAll() []*Task {
	return tm.tasks.Values()
}

func (tm *Manager) Get(tid uint64) (*Task, bool) {
	return tm.tasks.Load(tid)
}

func (tm *Manager) MustGet(tid uint64) *Task {
	task, _ := tm.Get(tid)
	return task
}

func (tm *Manager) Retry(tid uint64) error {
	t, ok := tm.Get(tid)
	if !ok {
		return ErrTaskNotFound
	}
	tm.do(t.ID)
	return nil
}

func (tm *Manager) Cancel(tid uint64) error {
	t, ok := tm.Get(tid)
	if !ok {
		return ErrTaskNotFound
	}
	t.Cancel()
	return nil
}

func (tm *Manager) Remove(tid uint64) {
	tm.tasks.Delete(tid)
}

func (tm *Manager) RemoveFinished() {
	tasks := tm.GetAll()
	for _, task := range tasks {
		if task.Status == FINISHED {
			tm.Remove(task.ID)
		}
	}
}

func (tm *Manager) RemoveError() {
	tasks := tm.GetAll()
	for _, task := range tasks {
		if task.Error != nil {
			tm.Remove(task.ID)
		}
	}
}

func NewTaskManager() *Manager {
	return &Manager{
		tasks: generic_sync.MapOf[uint64, *Task]{},
		curID: 0,
	}
}
