package task

import (
	"github.com/pkg/errors"
	"sync/atomic"

	"github.com/alist-org/alist/v3/pkg/generic_sync"
)

type Manager struct {
	works uint
	curID uint64
	tasks generic_sync.MapOf[uint64, *Task]
}

func (tm *Manager) Add(name string, f Func) uint64 {
	task := newTask(name, f)
	tm.addTask(task)
	go task.Run()
	return task.ID
}

func (tm *Manager) addTask(task *Task) {
	task.ID = tm.curID
	atomic.AddUint64(&tm.curID, 1)
	tm.tasks.Store(task.ID, task)
}

func (tm *Manager) GetAll() []*Task {
	return tm.tasks.Values()
}

func (tm *Manager) Get(id uint64) (*Task, bool) {
	return tm.tasks.Load(id)
}

func (tm *Manager) Retry(id uint64) error {
	t, ok := tm.Get(id)
	if !ok {
		return errors.New("task not found")
	}
	t.Retry()
	return nil
}

func (tm *Manager) Cancel(id uint64) error {
	t, ok := tm.Get(id)
	if !ok {
		return errors.New("task not found")
	}
	t.Cancel()
	return nil
}

func (tm *Manager) Remove(id uint64) {
	tm.tasks.Delete(id)
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
