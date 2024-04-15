package tache

// OnBeforeRetry is the interface for tasks that need to be executed before retrying
type OnBeforeRetry interface {
	OnBeforeRetry()
}

// OnSucceeded is the interface for tasks that need to be executed when they succeed
type OnSucceeded interface {
	OnSucceeded()
}

// OnFailed is the interface for tasks that need to be executed when they fail
type OnFailed interface {
	OnFailed()
}
