package aria2

import (
	"context"
	"github.com/alist-org/alist/v3/pkg/task"
)

func ListFinished(ctx context.Context) []*task.Task[string] {
	return DownTaskManager.GetByStates(task.Succeeded, task.CANCELED, task.ERRORED)
}

func ListUndone(ctx context.Context) []*task.Task[string] {
	return DownTaskManager.GetByStates(task.PENDING, task.RUNNING, task.CANCELING)
}
