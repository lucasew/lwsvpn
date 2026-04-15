package errors

import (
    "bytes"
    "errors"
    "log"
    "strings"
    "testing"
)

func TestReportError(t *testing.T) {
    var buf bytes.Buffer
    oldWriter := log.Writer()
    log.SetOutput(&buf)
    defer log.SetOutput(oldWriter) // Restore after test

    err := errors.New("test error")
    ReportError(err)

    if !strings.Contains(buf.String(), "[ERROR] test error") {
        t.Errorf("expected error message to contain '[ERROR] test error', got %q", buf.String())
    }
}
