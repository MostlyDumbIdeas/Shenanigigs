package errors

import (
	"fmt"

	goerrors "github.com/go-errors/errors"
)

type ErrorType string

const (
	ErrTypeNotFound     ErrorType = "NOT_FOUND"
	ErrTypeInvalidInput ErrorType = "INVALID_INPUT"
	ErrTypeUnauthorized ErrorType = "UNAUTHORIZED"
	ErrTypeInternal     ErrorType = "INTERNAL"
	ErrTypeUnavailable  ErrorType = "UNAVAILABLE"
	ErrTypeRateLimit    ErrorType = "RATE_LIMIT"
)

type DomainError struct {
	Type    ErrorType
	Message string
	Err     error
	Stack   []byte
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func (e *DomainError) StackTrace() []byte {
	return e.Stack
}

func New(errType ErrorType, message string, err error) *DomainError {
	var stack []byte
	if err != nil {
		if stackErr, ok := err.(*goerrors.Error); ok {
			stack = stackErr.Stack()
		} else {
			stack = goerrors.Wrap(err, 2).Stack()
		}
	} else {
		stack = goerrors.New(message).Stack()
	}

	return &DomainError{
		Type:    errType,
		Message: message,
		Err:     err,
		Stack:   stack,
	}
}

func NotFound(message string, err error) *DomainError {
	return New(ErrTypeNotFound, message, err)
}

func InvalidInput(message string, err error) *DomainError {
	return New(ErrTypeInvalidInput, message, err)
}

func Unauthorized(message string, err error) *DomainError {
	return New(ErrTypeUnauthorized, message, err)
}

func Internal(message string, err error) *DomainError {
	return New(ErrTypeInternal, message, err)
}

func Unavailable(message string, err error) *DomainError {
	return New(ErrTypeUnavailable, message, err)
}

func RateLimit(message string, err error) *DomainError {
	return New(ErrTypeRateLimit, message, err)
}
