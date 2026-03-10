package errors

import (
	"log"
)

// ReportError centralizes error reporting for the application.
// In the future, this can be wired to Sentry or another error tracking service.
func ReportError(err error, context string) {
	if err == nil {
		return
	}
	if context != "" {
		log.Printf("ERROR [%s]: %v", context, err)
	} else {
		log.Printf("ERROR: %v", err)
	}
}
