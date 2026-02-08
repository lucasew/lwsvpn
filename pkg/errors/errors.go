package errors

import (
    "fmt"
    "log/slog"
    "runtime/debug"
)

// ReportError logs the error with context and stack trace.
// It is the centralized place for error reporting.
// If Sentry is configured, this function should be updated to send the error to Sentry.
func ReportError(err error, context string) {
    if err == nil {
        return
    }

    // Log the error to stderr (standard log)
    slog.Error(context, "error", err, "stack", string(debug.Stack()))
}

// ReportPanic is a helper to recover from panics and report them.
func ReportPanic() {
    if r := recover(); r != nil {
        err, ok := r.(error)
        if !ok {
            err = fmt.Errorf("%v", r)
        }
        ReportError(err, "Panic recovered")
    }
}
