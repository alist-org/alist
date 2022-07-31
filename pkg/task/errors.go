package task

import "errors"

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrTaskRunning  = errors.New("task is running")
)
