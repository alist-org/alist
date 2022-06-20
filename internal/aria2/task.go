package aria2

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/pkg/task"
)

type Task struct {
	Account   driver.Driver
	ParentDir string
	T         task.Task
}
