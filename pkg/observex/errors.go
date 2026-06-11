package observex

import (
	"context"
	"errors"
)

// ErrorKind classifies observex errors into stable categories.
type ErrorKind string

const (
	// ErrorKindConfig identifies configuration errors.
	ErrorKindConfig ErrorKind = "config"
	// ErrorKindValidation identifies validation errors.
	ErrorKindValidation ErrorKind = "validation"
	// ErrorKindConnection identifies connection errors.
	ErrorKindConnection ErrorKind = "connection"
	// ErrorKindUnavailable identifies unavailable dependency errors.
	ErrorKindUnavailable ErrorKind = "unavailable"
	// ErrorKindTimeout identifies deadline and timeout errors.
	ErrorKindTimeout ErrorKind = "timeout"
	// ErrorKindAuth identifies authentication and authorization errors.
	ErrorKindAuth ErrorKind = "auth"
	// ErrorKindConflict identifies conflicting state errors.
	ErrorKindConflict ErrorKind = "conflict"
	// ErrorKindRateLimit identifies rate limit errors.
	ErrorKindRateLimit ErrorKind = "rate_limit"
	// ErrorKindCanceled identifies canceled operation errors.
	ErrorKindCanceled ErrorKind = "canceled"
	// ErrorKindNotFound identifies missing resource errors.
	ErrorKindNotFound ErrorKind = "not_found"
	// ErrorKindAlreadyExists identifies duplicate resource errors.
	ErrorKindAlreadyExists ErrorKind = "already_exists"
	// ErrorKindInternal identifies unexpected internal errors.
	ErrorKindInternal ErrorKind = "internal"
)

// Error is the structured error type returned by observex APIs.
type Error struct {
	// Kind classifies the error for retry and handling decisions.
	Kind ErrorKind
	// Op names the operation that produced the error.
	Op string
	// Message carries the caller-facing error detail.
	Message string
	// Cause preserves the wrapped error when one exists.
	Cause error
	// Retryable reports whether retrying the operation may succeed.
	Retryable bool
}

// NewError constructs an Error without a wrapped cause.
func NewError(kind ErrorKind, op string, message string, retryable bool) *Error {
	return newError(kind, op, message, retryable, nil)
}

// WrapError constructs an Error that unwraps to cause.
func WrapError(kind ErrorKind, op string, message string, retryable bool, cause error) *Error {
	return newError(kind, op, message, retryable, cause)
}

// Error formats the structured error.
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

// Unwrap returns the wrapped cause.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// IsKind reports whether err has the requested ErrorKind.
func IsKind(err error, kind ErrorKind) bool {
	var target *Error
	if errors.As(err, &target) {
		return target.Kind == kind
	}
	return false
}

// MapError converts common and foundationx errors into *Error.
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
	return ErrorKindInternal
}
