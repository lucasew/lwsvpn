package errors

import (
	"log"
)

// ReportError centralizes error reporting for the application.
func ReportError(err error) {
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
	}
}
