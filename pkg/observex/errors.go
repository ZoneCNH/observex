package observex

import (
	"context"
	"errors"

	"github.com/ZoneCNH/foundationx/pkg/foundationx"
)

// ErrorKind classifies observex errors using foundationx categories.
type ErrorKind = foundationx.ErrorKind

const (
	// ErrorKindConfig identifies configuration errors.
	ErrorKindConfig = foundationx.ErrorKindConfig
	// ErrorKindValidation identifies validation errors.
	ErrorKindValidation = foundationx.ErrorKindValidation
	// ErrorKindConnection identifies connection errors.
	ErrorKindConnection = foundationx.ErrorKindConnection
	// ErrorKindUnavailable identifies unavailable dependency errors.
	ErrorKindUnavailable = foundationx.ErrorKindUnavailable
	// ErrorKindTimeout identifies deadline and timeout errors.
	ErrorKindTimeout = foundationx.ErrorKindTimeout
	// ErrorKindAuth identifies authentication and authorization errors.
	ErrorKindAuth = foundationx.ErrorKindAuth
	// ErrorKindConflict identifies conflicting state errors.
	ErrorKindConflict = foundationx.ErrorKindConflict
	// ErrorKindRateLimit identifies rate limit errors.
	ErrorKindRateLimit = foundationx.ErrorKindRateLimit
	// ErrorKindCanceled identifies canceled operation errors.
	ErrorKindCanceled = foundationx.ErrorKindCanceled
	// ErrorKindNotFound identifies missing resource errors.
	ErrorKindNotFound = foundationx.ErrorKindNotFound
	// ErrorKindAlreadyExists identifies duplicate resource errors.
	ErrorKindAlreadyExists = foundationx.ErrorKindAlreadyExist
	// ErrorKindInternal identifies unexpected internal errors.
	ErrorKindInternal = foundationx.ErrorKindInternal
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
	return foundationx.IsKind(err, kind)
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
