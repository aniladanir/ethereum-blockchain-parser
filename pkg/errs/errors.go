package errs

import (
	"errors"
)

var (
	errAlreadyExist = &ErrorAlreadyExist{}
	errNotFound     = &ErrorNotFound{}
)

type ErrorNotFound struct {
}

func (err ErrorNotFound) Error() string {
	return "not found"
}

type ErrorAlreadyExist struct {
}

func (err ErrorAlreadyExist) Error() string {
	return "already exist"
}
func AlreadyExistErr() error {
	return errAlreadyExist
}

func NotFoundErr() error {
	return errNotFound
}

func IsAlreadyExistErr(err error) bool {
	return errors.Is(err, errAlreadyExist)
}

func IsNotFoundErr(err error) bool {
	return errors.Is(err, errNotFound)
}
