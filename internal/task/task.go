// manage task, such as file upload, file copy between accounts, offline download, etc.
package task

type Func func(task *Task) error

var (
	PENDING  = "pending"
	RUNNING  = "running"
	FINISHED = "finished"
)

type Task struct {
	ID     int64
	Name   string
	Status string
	Error  error
	Func   Func
}

func NewTask(name string, func_ Func) *Task {
	return &Task{
		Name:   name,
		Status: PENDING,
		Func:   func_,
	}
}

func (t *Task) SetStatus(status string) {
	t.Status = status
}

func (t *Task) Run() {
	t.Status = RUNNING
	t.Error = t.Func(t)
	t.Status = FINISHED
}
