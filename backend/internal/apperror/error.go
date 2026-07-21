package apperror

import (
	"errors"
	"fmt"
)

type Kind uint8

const (
	Validation Kind = iota + 1
	NotFound
	Conflict
)

type Error struct {
	Kind Kind
	Err  error
}

func (e *Error) Error() string { return e.Err.Error() }
func (e *Error) Unwrap() error { return e.Err }

func New(kind Kind, format string, args ...any) error {
	return &Error{Kind: kind, Err: fmt.Errorf(format, args...)}
}

func Wrap(kind Kind, err error) error {
	if err == nil {
		return nil
	}
	var existing *Error
	if errors.As(err, &existing) {
		return err
	}
	return &Error{Kind: kind, Err: err}
}

func KindOf(err error) (Kind, bool) {
	var applicationError *Error
	if !errors.As(err, &applicationError) {
		return 0, false
	}
	return applicationError.Kind, true
}
