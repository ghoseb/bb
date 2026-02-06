package cmdutil

import "fmt"

// ExitError carries an exit code and optional user-facing message.
type ExitError struct {
	Code int
	Msg  string
}

func (e *ExitError) Error() string {
	return e.Msg
}

// NewExitError creates an ExitError with the provided code and message.
func NewExitError(code int, msg string) *ExitError {
	return &ExitError{Code: code, Msg: msg}
}

// ValidationError captures invalid flag or argument usage.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return e.Msg
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}
