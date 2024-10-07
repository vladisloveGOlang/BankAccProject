package dto

import (
	"errors"
	"fmt"
)

type NotFoundError struct {
	Err error
}

func (e NotFoundError) Error() string {
	return e.Err.Error()
}

func (e NotFoundError) Unwrap() error { return e.Err }

func NotFoundErr(msg string) NotFoundError {
	return NotFoundError{Err: errors.New(msg)}
}

func NotFoundErrf(msg string, a ...interface{}) NotFoundError {
	return NotFoundError{Err: fmt.Errorf(msg, a...)}
}
