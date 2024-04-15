package tache

import "errors"

// TacheError is a custom error type
type TacheError struct {
	Msg string
}

func (e *TacheError) Error() string {
	return e.Msg
}

// NewErr creates a new TacheError
func NewErr(msg string) error {
	return &TacheError{Msg: msg}
}

//var (
//	ErrTaskNotFound = NewErr("task not found")
//	ErrTaskRunning  = NewErr("task is running")
//)

type unrecoverableError struct {
	error
}

func (e unrecoverableError) Unwrap() error {
	return e.error
}

// Unrecoverable wraps an error in `unrecoverableError` struct
func Unrecoverable(err error) error {
	return unrecoverableError{err}
}

// IsRecoverable checks if error is an instance of `unrecoverableError`
func IsRecoverable(err error) bool {
	return !errors.Is(err, unrecoverableError{})
}
