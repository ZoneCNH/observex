package observex

import (
	"context"
	"errors"

	"github.com/ZoneCNH/foundationx/pkg/foundationx"
)

type ErrorKind = foundationx.ErrorKind

const (
	ErrorKindConfig        = foundationx.ErrorKindConfig
	ErrorKindValidation    = foundationx.ErrorKindValidation
	ErrorKindConnection    = foundationx.ErrorKindConnection
	ErrorKindUnavailable   = foundationx.ErrorKindUnavailable
	ErrorKindTimeout       = foundationx.ErrorKindTimeout
	ErrorKindAuth          = foundationx.ErrorKindAuth
	ErrorKindConflict      = foundationx.ErrorKindConflict
	ErrorKindRateLimit     = foundationx.ErrorKindRateLimit
	ErrorKindCanceled      = foundationx.ErrorKindCanceled
	ErrorKindNotFound      = foundationx.ErrorKindNotFound
	ErrorKindAlreadyExists = foundationx.ErrorKindAlreadyExist
	ErrorKindInternal      = foundationx.ErrorKindInternal
)

type Error struct {
	Kind      ErrorKind
	Op        string
	Message   string
	Cause     error
	Retryable bool
}

func NewError(kind ErrorKind, op string, message string, retryable bool) *Error {
	return newError(kind, op, message, retryable, nil)
}

func WrapError(kind ErrorKind, op string, message string, retryable bool, cause error) *Error {
	return newError(kind, op, message, retryable, cause)
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	message := string(e.Kind)
	if e.Op != "" {
		message += ": " + e.Op
	}
	if e.Message != "" {
		message += ": " + e.Message
	}
	if e.Message == "" && e.Cause != nil {
		message += ": " + e.Cause.Error()
	}
	return message
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func IsKind(err error, kind ErrorKind) bool {
	var target *Error
	if errors.As(err, &target) {
		return target.Kind == kind
	}
	return foundationx.IsKind(err, kind)
}

func MapError(op string, err error) error {
	if err == nil {
		return nil
	}

	var target *Error
	if errors.As(err, &target) {
		return target
	}

	if errors.Is(err, context.Canceled) {
		return newError(ErrorKindCanceled, op, "", false, err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return newError(ErrorKindTimeout, op, "", true, err)
	}
	if foundationErr, ok := foundationx.AsFoundationError(err); ok {
		mappedOp := op
		if mappedOp == "" {
			mappedOp = foundationErr.Op
		}
		return newError(foundationErr.Kind, mappedOp, foundationErr.Message, foundationErr.Retryable, err)
	}

	return newError(ErrorKindInternal, op, err.Error(), false, err)
}

func newError(kind ErrorKind, op string, message string, retryable bool, cause error) *Error {
	if message == "" && cause != nil {
		message = cause.Error()
	}
	return &Error{
		Kind:      kind,
		Op:        op,
		Message:   message,
		Cause:     cause,
		Retryable: retryable,
	}
}

func validationError(op string, message string, cause error) *Error {
	return newError(ErrorKindValidation, op, message, false, cause)
}

func contextError(op string, cause error) *Error {
	kind := ErrorKindUnavailable
	retryable := false
	if errors.Is(cause, context.Canceled) {
		kind = ErrorKindCanceled
	} else if errors.Is(cause, context.DeadlineExceeded) {
		kind = ErrorKindTimeout
		retryable = true
	}
	return newError(kind, op, "", retryable, cause)
}

func errorKind(err error) ErrorKind {
	var target *Error
	if errors.As(err, &target) {
		return target.Kind
	}
	if foundationErr, ok := foundationx.AsFoundationError(err); ok {
		return foundationErr.Kind
	}
	return ErrorKindInternal
}
