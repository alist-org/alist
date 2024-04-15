package tache

import (
	"runtime"
	"sync"
	"time"
)

// sliceContains checks if a slice contains a value
func sliceContains[T comparable](slice []T, v T) bool {
	for _, vv := range slice {
		if vv == v {
			return true
		}
	}
	return false
}

// getCurrentGoroutineStack get current goroutine stack
func getCurrentGoroutineStack() string {
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// newDebounce returns a debounced function
func newDebounce(f func(), interval time.Duration) func() {
	var timer *time.Timer
	var lock sync.Mutex
	return func() {
		lock.Lock()
		defer lock.Unlock()
		if timer == nil {
			timer = time.AfterFunc(interval, f)
		} else {
			timer.Reset(interval)
		}
	}
}

// isRetry checks if a task is retry executed
func isRetry[T Task](task T) bool {
	return task.GetState() == StateWaitingRetry
}

// isLastRetry checks if a task is last retry
func isLastRetry[T Task](task T) bool {
	retry, maxRetry := task.GetRetry()
	return retry >= maxRetry
}

// needRetry judge whether the task need retry
func needRetry[T Task](task T) bool {
	// if task is not recoverable, return false
	if !IsRecoverable(task.GetErr()) {
		return false
	}
	// if task is not retryable, return false
	if r, ok := Task(task).(Retryable); ok && !r.Retryable() {
		return false
	}
	// only retry when task is errored or failed
	if sliceContains([]State{StateErrored, StateFailed}, task.GetState()) {
		retry, maxRetry := task.GetRetry()
		if retry < maxRetry {
			task.SetRetry(retry+1, maxRetry)
			task.SetState(StateWaitingRetry)
			return true
		}
	}
	return false
}
