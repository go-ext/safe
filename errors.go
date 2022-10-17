package syncsafe

import (
	"runtime/pprof"
	"strings"
)

// StackError specifies an object providing an error along with stack trace
type StackError interface {
	error
	StackTrace() string
}

// waitError provides an error keeping underlying error along with stack trace of all running goroutines on the moment
// of raising an error
type waitError struct {
	err   error
	stack string
}

// newWaitError returns a new waitError based on underlying err
func newWaitError(err error) waitError {
	sb := strings.Builder{}
	_ = pprof.Lookup("goroutine").WriteTo(&sb, 2)
	return waitError{
		err:   err,
		stack: sb.String(),
	}
}

// Error returns an underlying error text
func (e waitError) Error() string {
	return e.err.Error()
}

// StackTrace returns stack trace been created upon error created
func (e waitError) StackTrace() string {
	return e.stack
}
