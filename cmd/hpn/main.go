package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Exit codes: 0 success, 1 runtime error, 2 usage error (missing/invalid args).
const (
	exitSuccess = 0
	exitErr     = 1
	exitUsage   = 2
)

// ErrUsage marks usage errors (missing required flag, invalid args) so main can exit with code 2.
var ErrUsage = errors.New("usage")

// usageError wraps a user-facing message and unwraps to ErrUsage for exit code 2.
type usageError struct{ err error }

func (e *usageError) Error() string { return e.err.Error() }
func (e *usageError) Unwrap() error { return ErrUsage }

// usageErrorf returns an error that prints as msg but causes exit code 2.
func usageErrorf(format string, args ...interface{}) error {
	return &usageError{err: fmt.Errorf(format, args...)}
}

func exitCodeFor(err error) int {
	if err != nil && errors.Is(err, ErrUsage) {
		return exitUsage
	}
	return exitErr
}

func main() {
	// Create a context that is cancelled on SIGINT or SIGTERM.
	// This allows concurrent worker pools to drain cleanly instead of leaving
	// orphaned sub-processes when the user presses Ctrl+C.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitCodeFor(err))
	}
}
