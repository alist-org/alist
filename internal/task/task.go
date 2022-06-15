// manage task, such as file upload, file copy between accounts, offline download, etc.
package task

type Task struct {
	Name     string
	Status   string
	Error    error
	Finish   bool
	Children []*Task
}
