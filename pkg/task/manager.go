package task

import (
	"sync/atomic"

	"github.com/alist-org/alist/v3/pkg/generic_sync"
)

func NewTaskManager() *Manager {
	return &Manager{
		tasks: generic_sync.MapOf[int64, *Task]{},
		curID: 0,
	}
}

type Manager struct {
	curID int64
	tasks generic_sync.MapOf[int64, *Task]
}

func (tm *Manager) AddTask(task *Task) {
	task.ID = tm.curID
	atomic.AddInt64(&tm.curID, 1)
	tm.tasks.Store(task.ID, task)
}

func (tm *Manager) GetAll() []*Task {
	return tm.tasks.Values()
}

func (tm *Manager) Get(id int64) (*Task, bool) {
	return tm.tasks.Load(id)
}

func (tm *Manager) Remove(id int64) {
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

func (tm *Manager) Add(name string, f Func) {
	task := newTask(name, f)
	tm.AddTask(task)
	go task.Run()
}
