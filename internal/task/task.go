// manage task, such as file upload, file copy between accounts, offline download, etc.
package task

import "context"

type Task struct {
	Name     string
	Func     func(context.Context) error
	Status   string
	Error    error
	Progress int
}
