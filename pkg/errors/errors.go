package errors

import (
	"log"
)

// ReportError logs the error.
func ReportError(err error) {
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
}
