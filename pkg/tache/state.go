package tache

// State is the state of a task
type State int

const (
	// StatePending is the state of a task when it is pending
	StatePending = iota
	// StateRunning is the state of a task when it is running
	StateRunning
	// StateSucceeded is the state of a task when it succeeded
	StateSucceeded
	// StateCanceling is the state of a task when it is canceling
	StateCanceling
	// StateCanceled is the state of a task when it is canceled
	StateCanceled
	// StateErrored is the state of a task when it is errored (it will be retried)
	StateErrored
	// StateFailing is the state of a task when it is failing (executed OnFailed hook)
	StateFailing
	// StateFailed is the state of a task when it failed (no retry times left)
	StateFailed
	// StateWaitingRetry is the state of a task when it is waiting for retry
	StateWaitingRetry
	// StateBeforeRetry is the state of a task when it is executing OnBeforeRetry hook
	StateBeforeRetry
)
