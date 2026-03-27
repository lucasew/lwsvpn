package errors

import (
	"log"
)

// ReportError centralizes error logging and handling across the application.
// Following the Single Responsibility Principle, it decouples the reporting
// mechanism from business logic, making it easier to add Sentry or other
// reporting backends in the future.
func ReportError(err error) {
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}
}
