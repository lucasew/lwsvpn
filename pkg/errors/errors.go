package errors

import (
	"log"
	"os"
)

// ReportError logs an error centrally. This replaces raw print/log/sentry calls.
func ReportError(err error, context string) {
	if err != nil {
		log.Printf("[ERROR] %s: %v", context, err)
	}
}

// ReportFatal logs an error centrally and terminates the program. This replaces panic.
func ReportFatal(err error, context string) {
	if err != nil {
		log.Printf("[FATAL] %s: %v", context, err)
		os.Exit(1)
	}
}
