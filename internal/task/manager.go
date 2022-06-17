package task

import (
	"sync/atomic"

	"github.com/alist-org/alist/v3/pkg/generic_sync"
)

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: generic_sync.MapOf[int64, *Task]{},
		curID: 0,
	}
}

type TaskManager struct {
	curID int64
	tasks generic_sync.MapOf[int64, *Task]
}

func (tm *TaskManager) AddTask(task *Task) {
	task.ID = tm.curID
	atomic.AddInt64(&tm.curID, 1)
	tm.tasks.Store(task.ID, task)
}

func (tm *TaskManager) GetAll() []*Task {
	return tm.tasks.Values()
}

func (tm *TaskManager) Get(id int64) (*Task, bool) {
	return tm.tasks.Load(id)
}

func (tm *TaskManager) Remove(id int64) {
	tm.tasks.Delete(id)
}

func (tm *TaskManager) RemoveFinished() {
	tasks := tm.GetAll()
	for _, task := range tasks {
		if task.Status == FINISHED {
			tm.Remove(task.ID)
		}
	}
}

func (tm *TaskManager) RemoveError() {
	tasks := tm.GetAll()
	for _, task := range tasks {
		if task.Error != nil {
			tm.Remove(task.ID)
		}
	}
}

func (tm *TaskManager) Add(name string, f Func) {
	task := NewTask(name, f)
	tm.AddTask(task)
	go task.Run()
}
