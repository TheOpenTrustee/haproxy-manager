package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

var Join = errors.Join

// CustomError wraps an error with a stack trace
type CustomError struct {
	Err        error
	StackTrace string
}

// Error implements the error interface
func (e *CustomError) Error() string {
	return e.Err.Error()
}

// Unwrap allows errors.Is and errors.As to work with CustomError
func (e *CustomError) Unwrap() error {
	return e.Err
}

// New creates a new CustomError with a stack trace
func New(msg string) error {
	return &CustomError{
		Err:        fmt.Errorf("%s", msg),
		StackTrace: CaptureStackTrace(3),
	}
}

// captureStackTrace captures the call stack with indentation
func CaptureStackTrace(skip int) string {
	var sb strings.Builder
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip, pcs)
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	return sb.String()
}
