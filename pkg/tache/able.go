package tache

// Persistable judge whether the task is persistable
type Persistable interface {
	Persistable() bool
}

// Recoverable judge whether the task is recoverable
type Recoverable interface {
	Recoverable() bool
}

// Retryable judge whether the task is retryable
type Retryable interface {
	Retryable() bool
}
